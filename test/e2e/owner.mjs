// Acts as the owner in the built-in UI for the data path test (datapath_test.go), the way docs/dev/testing.md asks:
// Chromium at http://localhost through kubectl port-forward to the core (D-80, D-81), a passkey created through a
// one-time enrollment link with Chromium's virtual authenticator, and no flag that weakens the product's origin check.
// It then sends the owner's forms from the signed-in page itself, as the UI's own forms send them: the browser adds
// the session cookie and the same-origin headers, so the UI and the core check them as they check the owner's.
// Usage: OPENFIEF_E2E_ENROLL_LINK=<one-time enrollment link> node owner.mjs
// Protocol: once signed in it prints one line, "owner: ready". It then reads one JSON request per line on stdin,
//   {"path": "/accounts", "form": {"name": "..."}}                       a form, sent URL-encoded
//   {"path": "/accounts/x/statements", "multipart": [{"name": "mode", "value": "auto"},
//    {"name": "statement", "filename": "f.xlsx", "base64": "..."}]}      a multipart form, in the order given
// POSTs it, follows redirects, and prints one line per request: "owner-result: " and the JSON
// {"status": 200, "redirected": true, "path": "/accounts", "body": "..."}. A request
//   {"page": "/?date=2026-03-13"}                                         a page to look at, with its charts
// instead opens the page in the signed-in tab, waits until app.js has drawn every chart on it, and answers
// {"status": 200, "path": "/?date=...", "charts": [...], "bars": [...]}: per canvas what its Chart.js chart draws,
// its days and per dataset its values ("" for a gap), and per bar of the allocation its key, share, width and value
// as the page writes them (charts_test.go). It exits when stdin ends, non-zero on the first failure to sign in or to
// send. It never prints the session cookie.
import { existsSync, readdirSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import readline from 'node:readline';
import { chromium } from 'playwright-core';

const link = process.env.OPENFIEF_E2E_ENROLL_LINK;
if (!link || !link.includes('/enroll#code=')) {
  console.error('usage: OPENFIEF_E2E_ENROLL_LINK=<enrollment link> node owner.mjs');
  process.exit(2);
}
const origin = new URL(link).origin;
const timeout = 60_000;
// maxBody bounds the page text sent back for each request.
const maxBody = 256 * 1024;

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

// send POSTs one form from the signed-in page, as a form the UI serves would, and returns what came back.
async function send(page, req) {
  return page.evaluate(
    async ({ req, maxBody }) => {
      let body;
      if (req.multipart) {
        body = new FormData();
        for (const part of req.multipart) {
          if (part.filename) {
            const bytes = Uint8Array.from(atob(part.base64), (c) => c.charCodeAt(0));
            body.append(part.name, new File([bytes], part.filename, { type: 'application/octet-stream' }));
          } else {
            body.append(part.name, part.value);
          }
        }
      } else {
        body = new URLSearchParams(req.form || {});
      }
      const res = await fetch(req.path, { method: 'POST', body, credentials: 'same-origin', redirect: 'follow' });
      const text = await res.text();
      return { status: res.status, redirected: res.redirected, path: new URL(res.url).pathname, body: text.slice(0, maxBody) };
    },
    { req, maxBody },
  );
}

// look opens path in the signed-in tab and reads what its charts draw: what Chart.js holds for each canvas once
// app.js has drawn it, not the data attribute it was drawn from, and each allocation bar's text and width.
async function look(page, path) {
  const res = await page.goto(origin + path, { timeout });
  await page.waitForFunction(
    () => Array.from(document.querySelectorAll('canvas[data-chart]')).every((c) => c.dataset.drawn === 'true'),
    null,
    { timeout },
  );
  const drawn = await page.evaluate(() => {
    const text = (el) => (el ? el.textContent.trim() : '');
    const charts = Array.from(document.querySelectorAll('canvas[data-chart]')).map((canvas) => {
      const chart = window.Chart.getChart(canvas);
      return {
        aria: canvas.getAttribute('aria-label'),
        labels: chart.data.labels.map(String),
        series: chart.data.datasets.map((d) => d.data.map((v) => (v === null || v === undefined ? '' : String(v)))),
      };
    });
    const bars = Array.from(document.querySelectorAll('ul.bars > li')).map((li) => ({
      key: text(li.querySelector('.bar-head > span:first-child')),
      share: text(li.querySelector('.bar-head > .mono')),
      width: (li.querySelector('rect.fill') || { getAttribute: () => '' }).getAttribute('width') || '',
      value: text(li.querySelector('.bar-value')),
    }));
    return { path: location.pathname + location.search, charts, bars };
  });
  return { status: res ? res.status() : 0, ...drawn };
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
  await page.fill('#passkey-name', 'e2e-data-path');
  await page.click('[data-webauthn="register"]');
  await page.waitForURL((url) => url.origin === origin && !url.pathname.startsWith('/enroll'), { timeout });
  await page.getByRole('button', { name: 'Sign out' }).waitFor({ timeout });
  console.log('owner: ready');
  const lines = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });
  for await (const line of lines) {
    if (!line.trim()) {
      continue;
    }
    const req = JSON.parse(line);
    const result = req.page ? await look(page, req.page) : await send(page, req);
    console.log('owner-result: ' + JSON.stringify(result));
  }
} catch (err) {
  failed = true;
  console.error(`owner: ${err && err.stack ? err.stack : err}`);
} finally {
  await browser.close();
}
process.exit(failed ? 1 : 0);
