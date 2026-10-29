// Drives a real browser through the owner's first passkey and a passkey login, the way docs/dev/testing.md asks:
// Chromium at http://localhost through kubectl port-forward to the core (D-80, D-81), with Chromium's virtual
// authenticator, and no flag that weakens the product's origin check. passkey_test.go runs it.
// Usage: OPENFIEF_E2E_ENROLL_LINK=<one-time enrollment link> node passkey.mjs
// It prints one line per step and exits non-zero on the first failure.
import { existsSync, readdirSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { chromium } from 'playwright-core';

const link = process.env.OPENFIEF_E2E_ENROLL_LINK;
if (!link || !link.includes('/enroll#code=')) {
  console.error('usage: OPENFIEF_E2E_ENROLL_LINK=<enrollment link> node passkey.mjs');
  process.exit(2);
}
const origin = new URL(link).origin;
const timeout = 30_000;

// browserOptions picks the Chromium to drive without downloading one, as scripts/docs-diagrams.mjs does.
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

// expectHome waits for the overview at the origin's root and checks it is the owner's.
async function expectHome(page, step) {
  await page.waitForURL(origin + '/', { timeout });
  const signOut = page.getByRole('button', { name: 'Sign out' });
  await signOut.waitFor({ timeout });
  console.log(`passkey: ${step}: on ${page.url()}, signed in`);
}

// sessionCookie returns the session cookie the browser keeps for the origin.
async function sessionCookie(context) {
  const cookies = await context.cookies(origin);
  return cookies.find((c) => c.name === '__Host-openfief_session');
}

const browser = await chromium.launch({ headless: true, ...browserOptions() });
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

  await page.goto(link, { timeout });
  await page.fill('#passkey-name', 'e2e-virtual');
  await page.click('[data-webauthn="register"]');
  await expectHome(page, 'created the first passkey');
  const cookie = await sessionCookie(context);
  if (!cookie || !cookie.secure || !cookie.httpOnly || cookie.sameSite !== 'Strict' || cookie.path !== '/') {
    throw new Error(`the browser kept no __Host- session cookie on ${origin}: ${JSON.stringify(cookie)}`);
  }
  console.log(`passkey: the browser keeps __Host-openfief_session on ${origin} (D-85)`);

  await page.getByRole('button', { name: 'Sign out' }).click();
  await page.waitForURL(origin + '/login', { timeout });
  if (await sessionCookie(context)) {
    throw new Error('the session cookie survived signing out');
  }
  console.log('passkey: signed out');

  await page.click('[data-webauthn="login"]');
  await expectHome(page, 'signed in with the passkey');
} catch (err) {
  failed = true;
  console.error(`passkey: ${err && err.stack ? err.stack : err}`);
} finally {
  await browser.close();
}
process.exit(failed ? 1 : 0);
