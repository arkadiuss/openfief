// Drives a real browser through D-200's criterion 8 for a second hostname, as the owner does it from the settings page:
// a passkey on http://localhost (through kubectl port-forward to the core, D-80), then, signed in there, Settings' "Passkey
// on another host" form creates a one-time enrollment link for the second origin (D-60, D-183), and on that origin the same
// browser creates a passkey with the link, keeps the __Host- session cookie, signs out and signs in again with it.
// settingslink_test.go runs it once the core accepts the second origin.
// Usage: OPENFIEF_E2E_ENROLL_LINK=<one-time link for localhost> OPENFIEF_E2E_SECOND_ORIGIN=https://<hostname>
//        OPENFIEF_E2E_HOST_RULES=... OPENFIEF_E2E_TRUSTED_SPKI=... node secondhost.mjs
// OPENFIEF_E2E_HOST_RULES maps the hostname to where the test forwards the Gateway (Chromium's --host-resolver-rules), and
// OPENFIEF_E2E_TRUSTED_SPKI lists the keys of the run's CA and the Gateway's certificate, the only ones Chromium then
// accepts besides its own roots (--ignore-certificate-errors-spki-list), as passkey.mjs does. No flag weakens the
// product's origin check. It prints one line per step and exits non-zero on the first failure; it never prints a link's
// code or the session cookie.
import { existsSync, readdirSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { chromium } from 'playwright-core';

const first = process.env.OPENFIEF_E2E_ENROLL_LINK;
const second = process.env.OPENFIEF_E2E_SECOND_ORIGIN;
if (!first || !first.includes('/enroll#code=') || !second || !second.startsWith('https://')) {
  console.error('usage: OPENFIEF_E2E_ENROLL_LINK=<link> OPENFIEF_E2E_SECOND_ORIGIN=https://<hostname> node secondhost.mjs');
  process.exit(2);
}
const firstOrigin = new URL(first).origin;
const secondHost = new URL(second).hostname;
const timeout = 30_000;

// browserOptions picks the Chromium to drive without downloading one, as passkey.mjs does.
function browserOptions() {
  if (process.env.E2E_CHROMIUM) {
    return { executablePath: process.env.E2E_CHROMIUM };
  }
  const roots = [process.env.PLAYWRIGHT_BROWSERS_PATH, '/opt/pw-browsers', path.join(os.homedir(), '.cache', 'ms-playwright')];
  for (const root of roots.filter((r) => r && existsSync(r))) {
    for (const dir of readdirSync(root).filter((d) => /^chromium-\d+$/.test(d)).sort().reverse()) {
      for (const bin of ['chrome-linux64/chrome', 'chrome-linux/chrome', 'chrome-mac/Chromium.app/Contents/MacOS/Chromium']) {
        const candidate = path.join(root, dir, bin);
        if (existsSync(candidate)) {
          return { executablePath: candidate };
        }
      }
    }
  }
  return { channel: 'chrome' };
}

// chromiumArgs points Chromium at the Gateway for the second hostname and trusts the run's keys.
function chromiumArgs() {
  const args = [];
  if (process.env.OPENFIEF_E2E_HOST_RULES) {
    args.push(`--host-resolver-rules=${process.env.OPENFIEF_E2E_HOST_RULES}`);
  }
  if (process.env.OPENFIEF_E2E_TRUSTED_SPKI) {
    args.push(`--ignore-certificate-errors-spki-list=${process.env.OPENFIEF_E2E_TRUSTED_SPKI}`);
  }
  return args;
}

// expectHome waits for the owner's landing page on origin, the overview or the first-run wizard (D-73), signed in.
async function expectHome(page, origin, step) {
  await page.waitForURL((url) => url.origin === origin && (url.pathname === '/' || url.pathname === '/setup'), {
    timeout,
  });
  await page.getByRole('button', { name: 'Sign out' }).waitFor({ timeout });
  console.log(`secondhost: ${step}: on ${new URL(page.url()).origin}${new URL(page.url()).pathname}, signed in`);
}

// createPasskey opens an enrollment link, names the passkey and creates it with the virtual authenticator.
async function createPasskey(page, link, name) {
  await page.goto(link, { timeout });
  await page.fill('#passkey-name', name);
  await page.click('[data-webauthn="register"]');
}

// enrollmentLink creates, from the settings page on the first origin, a one-time link for the second origin. Creating
// one needs a passkey check in the last five minutes (D-183): when the page asks for it, it confirms with the passkey
// and the page posts the form again.
async function enrollmentLink(page) {
  await page.goto(`${firstOrigin}/settings`, { timeout });
  await page.fill('#target-origin', second);
  await Promise.all([
    page.waitForURL(`${firstOrigin}/settings/enrollment-links`, { timeout }),
    page.click('form[action="/settings/enrollment-links"] button[type="submit"]'),
  ]);
  if (await page.locator('#passkey-check-retry').count()) {
    console.log('secondhost: the settings page asks for a passkey check first (D-183); confirming');
    // The page is already at the refused post's answer, so wait for the post again, not the URL.
    const posted = page.waitForResponse(
      (r) => r.url() === `${firstOrigin}/settings/enrollment-links` && r.request().method() === 'POST',
      { timeout },
    );
    await page.click('[data-webauthn="check"]');
    await posted;
  }
  const code = page.locator('section:has(#link-title) code.secret');
  await code.waitFor({ timeout });
  const url = (await code.textContent()).trim();
  if (!url.startsWith(`${second}/enroll#code=`)) {
    throw new Error(`the settings page made a link for ${new URL(url).origin}, not ${second}`);
  }
  console.log(`secondhost: the settings page created a one-time link for ${second}`);
  return url;
}

// sessionCookie returns the session cookie the browser keeps for origin.
async function sessionCookie(context, origin) {
  const cookies = await context.cookies(origin);
  return cookies.find((c) => c.name === '__Host-openfief_session');
}

// passkeyHosts lists the hosts of the passkeys the settings page on origin shows.
async function passkeyHosts(page, origin) {
  await page.goto(`${origin}/settings`, { timeout });
  const cells = page.locator('section:has(#passkeys-title) tbody tr td:nth-child(2)');
  return (await cells.allTextContents()).map((t) => t.trim());
}

const browser = await chromium.launch({ headless: true, args: chromiumArgs(), ...browserOptions() });
let failed = false;
try {
  const context = await browser.newContext();
  const page = await context.newPage();
  const cdp = await context.newCDPSession(page);
  await cdp.send('WebAuthn.enable');
  await cdp.send('WebAuthn.addVirtualAuthenticator', {
    options: {
      protocol: 'ctap2',
      transport: 'internal',
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });

  await createPasskey(page, first, 'e2e-first-host');
  await expectHome(page, firstOrigin, 'created a passkey on the first host');
  const link = await enrollmentLink(page);

  await createPasskey(page, link, 'e2e-second-host');
  await expectHome(page, second, 'created a passkey on the second host with the settings page link');
  const cookie = await sessionCookie(context, second);
  if (!cookie || !cookie.secure || !cookie.httpOnly || cookie.sameSite !== 'Strict' || cookie.path !== '/') {
    throw new Error(`the browser kept no __Host- session cookie on ${second}`);
  }
  console.log(`secondhost: the browser keeps __Host-openfief_session on ${second} (D-85)`);

  await page.getByRole('button', { name: 'Sign out' }).click();
  await page.waitForURL(`${second}/login`, { timeout });
  if (await sessionCookie(context, second)) {
    throw new Error(`the session cookie on ${second} survived signing out`);
  }
  console.log(`secondhost: signed out of ${second}`);
  await page.click('[data-webauthn="login"]');
  await expectHome(page, second, 'signed in on the second host with its passkey');

  const hosts = await passkeyHosts(page, second);
  for (const want of ['localhost', secondHost]) {
    if (!hosts.includes(want)) {
      throw new Error(`the settings page lists passkeys on ${JSON.stringify(hosts)}, none on ${want}`);
    }
  }
  console.log(`secondhost: the settings page lists passkeys on ${hosts.join(', ')}`);
} catch (err) {
  failed = true;
  console.error(`secondhost: ${err && err.stack ? err.stack : err}`);
} finally {
  await browser.close();
}
process.exit(failed ? 1 : 0);
