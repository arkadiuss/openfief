#!/usr/bin/env node
// Browser checks of the built-in UI over its in-memory fakes (task ui:browser). It starts them itself with
// TestServeForBrowser (go test) on ADDR, default localhost:8099, and stops them when done: a screenshot of every main page at 390 px in light and dark and at desktop width; no page
// scrolls sideways at 360 or 390 px; the menu folds the navigation away on a phone; the manifest makes the UI
// installable; the service worker caches only static files, serves the offline page instead of a page when the
// network is gone, and empties its cache on sign-out.
// Usage: node scripts/ui-browser.mjs <screenshot-dir>
// The browser is DOCS_CHROMIUM, else a Playwright-managed Chromium, else the installed Google Chrome.
import { existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync } from 'node:fs';
import { spawn } from 'node:child_process';
import os from 'node:os';
import path from 'node:path';
import { chromium } from 'playwright-core';

const [out] = process.argv.slice(2);
if (!out) {
  console.error('usage: node scripts/ui-browser.mjs <screenshot-dir>');
  process.exit(2);
}
mkdirSync(out, { recursive: true });
const addr = process.env.ADDR || 'localhost:8099';
const base = `http://${addr}`;

// serve starts the UI over its fakes in a process group of its own and waits until it answers.
async function serve() {
  const server = spawn('go', ['test', './internal/views/ui/', '-run', '^TestServeForBrowser$', '-count=1', '-timeout', '15m'], {
    env: { ...process.env, OPENFIEF_UI_SERVE: addr }, stdio: ['ignore', 'inherit', 'inherit'], detached: true,
  });
  for (let i = 0; i < 120; i++) {
    try {
      if ((await fetch(base + '/healthz')).ok) {
        return server;
      }
    } catch {
      // not listening yet
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
  process.kill(-server.pid);
  throw new Error(`the UI over its fakes did not answer on ${base}`);
}

// The fakes' one account is acc-synthetic-1 (internal/views/ui/fakes_test.go).
const pages = {
  overview: '/', accounts: '/accounts', account: '/accounts/acc-synthetic-1', assertions: '/assertions',
  'price-checks': '/price-checks', transfers: '/transfers', sleeves: '/sleeves', plugins: '/plugins', settings: '/settings', setup: '/setup',
};
const signedOut = { login: '/login', enroll: '/enroll', offline: '/static/offline.html' };

// browserOptions picks the Chromium to drive, without downloading one (as docs-diagrams.mjs does).
function browserOptions() {
  if (process.env.DOCS_CHROMIUM) {
    return { executablePath: process.env.DOCS_CHROMIUM };
  }
  const roots = [process.env.PLAYWRIGHT_BROWSERS_PATH, '/opt/pw-browsers', path.join(os.homedir(), '.cache', 'ms-playwright')];
  for (const dir of roots.filter((r) => r && existsSync(r))) {
    for (const sub of readdirSync(dir).filter((d) => /^chromium-\d+$/.test(d)).sort().reverse()) {
      for (const bin of ['chrome-linux64/chrome', 'chrome-linux/chrome', 'chrome-mac/Chromium.app/Contents/MacOS/Chromium']) {
        const candidate = path.join(dir, sub, bin);
        if (existsSync(candidate)) {
          return { executablePath: candidate };
        }
      }
    }
  }
  return { channel: 'chrome' };
}

const failures = [];
function check(ok, message) {
  if (!ok) {
    failures.push(message);
    console.error(`FAIL ${message}`);
  }
}

async function context(browser, options, signedIn) {
  const ctx = await browser.newContext(options);
  if (signedIn) {
    // The fakes' server sets the owner's session cookie here (TestServeForBrowser), as a sign-in does.
    const page = await ctx.newPage();
    await page.goto(base + '/test/sign-in');
    await page.close();
  }
  return ctx;
}

// shoot opens every page in a context and screenshots it, checking that none scrolls sideways.
async function shoot(browser, label, options) {
  for (const [list, signedIn] of [[pages, true], [signedOut, false]]) {
    const ctx = await context(browser, options, signedIn);
    const page = await ctx.newPage();
    for (const [name, target] of Object.entries(list)) {
      const response = await page.goto(base + target, { waitUntil: 'networkidle' });
      check(response && response.status() < 400, `${label} ${name}: status ${response && response.status()}`);
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
      check(overflow <= 0, `${label} ${name}: scrolls ${overflow} px sideways`);
      await page.screenshot({ path: path.join(out, `${label}-${name}.png`), fullPage: true });
    }
    await ctx.close();
  }
}

// phoneRows checks that rows an htmx swap adds to a table are labelled like the first ones: the account's next page
// of transactions replaces the "More transactions" row.
async function phoneRows(browser) {
  const ctx = await context(browser, { viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  await page.goto(base + '/accounts/acc-synthetic-1', { waitUntil: 'networkidle' });
  const more = page.locator('a', { hasText: 'More transactions' });
  if (await more.count() > 0) {
    const rows = await page.locator('.table.stack tbody tr').count();
    await more.first().click();
    // Wait for the swap to add rows, then for every row to carry its labels.
    await page.waitForFunction((n) => document.querySelectorAll('.table.stack tbody tr').length > n, rows);
    const labelled = await page.waitForFunction(
      () => document.querySelectorAll('.table.stack tbody tr:not([data-labelled])').length === 0, null, { timeout: 5000 },
    ).then(() => true, () => false);
    check(labelled, 'phone: rows swapped into a table carry no labels');
  } else {
    console.log('note: the fakes have no second page of transactions; swapped rows not checked');
  }
  await ctx.close();
}

// phoneMenu checks that the navigation is folded away on a phone and the menu button opens it.
async function phoneMenu(browser) {
  const ctx = await context(browser, { viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  await page.goto(base + '/', { waitUntil: 'networkidle' });
  check(!(await page.isVisible('nav[aria-label="Main"]')), 'phone: navigation shown before the menu opens');
  await page.click('.menu-toggle');
  check(await page.isVisible('nav[aria-label="Main"] a[href="/settings"]'), 'phone: menu button does not open the navigation');
  check((await page.getAttribute('.menu-toggle', 'aria-expanded')) === 'true', 'phone: aria-expanded not true when open');
  await page.screenshot({ path: path.join(out, 'phone-light-menu-open.png') });
  await ctx.close();
}

// app checks installability, the worker's cache, the offline page and sign-out, in a profile of its own: an
// incognito context is never installable.
async function app() {
  const profile = mkdtempSync(path.join(os.tmpdir(), 'openfief-ui-'));
  const ctx = await chromium.launchPersistentContext(profile, { ...browserOptions(), viewport: { width: 1280, height: 800 } });
  const page = ctx.pages()[0] || await ctx.newPage();
  await page.goto(base + '/test/sign-in', { waitUntil: 'networkidle' });
  await page.goto(base + '/', { waitUntil: 'networkidle' });
  await page.evaluate(() => navigator.serviceWorker.ready);
  await page.reload({ waitUntil: 'networkidle' });
  const cdp = await ctx.newCDPSession(page);
  const { installabilityErrors } = await cdp.send('Page.getInstallabilityErrors');
  check(installabilityErrors.length === 0, `not installable: ${JSON.stringify(installabilityErrors)}`);
  const manifest = await cdp.send('Page.getAppManifest');
  check(manifest.errors.length === 0, `manifest errors: ${JSON.stringify(manifest.errors)}`);
  check(await page.evaluate(() => navigator.serviceWorker.controller !== null), 'the service worker does not control the page');
  const cached = await page.evaluate(async () => {
    const urls = [];
    for (const name of await caches.keys()) {
      for (const request of await (await caches.open(name)).keys()) {
        urls.push(new URL(request.url).pathname);
      }
    }
    return urls;
  });
  check(cached.length > 0, 'the service worker cached nothing');
  for (const url of cached) {
    check(url.startsWith('/static/') || url === '/manifest.webmanifest', `the service worker cached ${url}`);
  }
  await ctx.setOffline(true);
  await page.goto(base + '/accounts', { waitUntil: 'load' }).catch(() => {});
  const offline = await page.textContent('body');
  check(offline.includes("can't be reached") && !offline.includes('Synthetic brokerage'), 'offline: a page instead of the offline page');
  await page.screenshot({ path: path.join(out, 'desktop-light-offline-navigation.png') });
  await ctx.setOffline(false);
  await page.goto(base + '/', { waitUntil: 'networkidle' });
  await Promise.all([page.waitForURL('**/login'), page.click('form[action="/logout"] button')]);
  const left = await page.evaluate(async () => (await caches.keys()).length);
  check(left === 0, `sign-out left ${left} caches`);
  await ctx.close();
  rmSync(profile, { recursive: true, force: true });
}

const server = await serve();
const browser = await chromium.launch(browserOptions());
try {
  for (const scheme of ['light', 'dark']) {
    await shoot(browser, `phone-${scheme}`, { viewport: { width: 390, height: 844 }, colorScheme: scheme, isMobile: true, hasTouch: true });
  }
  await shoot(browser, 'phone360-light', { viewport: { width: 360, height: 780 }, colorScheme: 'light', isMobile: true, hasTouch: true });
  await shoot(browser, 'desktop-light', { viewport: { width: 1280, height: 800 }, colorScheme: 'light' });
  await phoneMenu(browser);
  await phoneRows(browser);
  await app();
} finally {
  await browser.close();
  process.kill(-server.pid);
}
if (failures.length > 0) {
  console.error(`${failures.length} browser check(s) failed`);
  process.exit(1);
}
console.log(`browser checks passed; screenshots in ${out}`);
