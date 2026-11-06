#!/usr/bin/env node
// Browser checks of the built-in UI over its in-memory fakes (task ui:browser). It starts them itself with
// TestServeForBrowser (go test) on ADDR, default localhost:8099, and stops them when done: a screenshot of every main page at 390 px in light and dark and at desktop width; no page
// scrolls sideways at 360 or 390 px; the menu folds the navigation away on a phone, with the Inbox's open count; the check
// pages D-192 removed land on the Inbox at 360 px with its in-place forms, and an account's banner links there; the manifest makes the UI
// installable; the service worker caches only static files, serves the offline page instead of a page when the
// network is gone, and empties its cache on sign-out; an action the fake core refuses for want of a recent passkey
// check (D-183) asks to confirm with a passkey, runs the check with Chromium's virtual authenticator, and goes ahead;
// at 360 px the overview's Ask box (D-194) takes "/", answers a suggestion chip, a typed command and one it doesn't
// understand, its chart card switches to the performance, and the Inbox shows its side column of Ask suggestions;
// at 360 px an account's tabs wrap and switch in place (D-195), an instrument's page shows its parts and takes the
// owner's rename, and the overview's chart card switches to the allocation, whose currency bar opens its holdings; the overview is kept offline only
// encrypted, under a key that can't be exported, shown offline marked as a copy, and wiped on sign-out and when the
// session ends (D-151).
// Usage: node scripts/ui-browser.mjs <screenshot-dir>
// The browser is DOCS_CHROMIUM, else a Playwright-managed Chromium, else the installed Google Chrome.
import { existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync } from 'node:fs';
import { spawn } from 'node:child_process';
import { generateKeyPairSync, randomBytes } from 'node:crypto';
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
  overview: '/', 'overview-ask': '/?q=what+needs+me', 'overview-performance': '/?over=performance', 'overview-allocation': '/?over=allocation',
  accounts: '/accounts', account: '/accounts/acc-synthetic-1', 'account-transactions': '/accounts/acc-synthetic-1?tab=transactions',
  'account-lots': '/accounts/acc-synthetic-1?tab=lots', 'account-cash': '/accounts/acc-synthetic-1?tab=cash',
  'account-reconciliation': '/accounts/acc-synthetic-1?tab=reconciliation', 'account-settings': '/accounts/acc-synthetic-1?tab=settings',
  instrument: '/instruments/XX0000000001?account=acc-synthetic-1', holdings: '/holdings?by=currency&key=EUR', inbox: '/inbox',
  'inbox-acknowledged': '/inbox?view=acknowledged', 'inbox-done': '/inbox?view=done', sleeves: '/sleeves', plugins: '/plugins',
  settings: '/settings', setup: '/setup',
};
// The check pages D-192 removed; each redirects to the Inbox.
const removedPages = ['/attention', '/tickers', '/assertions', '/price-checks', '/transfers'];
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
  await page.goto(base + '/accounts/acc-synthetic-1?tab=transactions', { waitUntil: 'networkidle' });
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
  check(await page.isVisible('nav[aria-label="Main"] a[href="/inbox"] .pill-gold'), 'phone: the menu shows no Inbox with its open count');
  check((await page.getAttribute('.menu-toggle', 'aria-expanded')) === 'true', 'phone: aria-expanded not true when open');
  await page.screenshot({ path: path.join(out, 'phone-light-menu-open.png') });
  await ctx.close();
}

// removedRedirect checks at phone width that each check page D-192 removed lands on the Inbox, which neither
// scrolls sideways nor hides its open items' forms, and that an account's banner links its own items there.
async function removedRedirect(browser) {
  const ctx = await context(browser, { viewport: { width: 360, height: 780 }, isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  for (const target of removedPages) {
    const response = await page.goto(base + target, { waitUntil: 'networkidle' });
    check(new URL(page.url()).pathname === '/inbox', `phone: ${target} lands on ${page.url()}, not the Inbox`);
    check(response && response.status() === 200, `phone: ${target} ends with status ${response && response.status()}`);
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    check(overflow <= 0, `phone: the Inbox from ${target} scrolls ${overflow} px sideways`);
  }
  for (const form of ['/inbox/tickers/confirm', '/inbox/retry-sync', '/accounts/bind', '/inbox/price-checks/acknowledge']) {
    check(await page.locator(`form[action="${form}"]`).count() > 0, `phone: the Inbox has no ${form} form`);
  }
  await page.goto(base + '/accounts/acc-synthetic-1', { waitUntil: 'networkidle' });
  check(await page.isVisible('.banner a[href^="/inbox#item-"]'), "phone: the account's banner doesn't link its items");
  await page.screenshot({ path: path.join(out, 'phone360-light-account-banner.png') });
  await ctx.close();
}

// askBox checks the overview's Ask box at phone width (D-194): "/" focuses it, a suggestion chip and a typed command
// are answered in place as cards with an Open link, a question it doesn't understand gets the commands it does, the
// chart card switches to the performance, and none of it scrolls sideways; then the Inbox's side column of Ask
// suggestions shows and links to the overview.
async function askBox(browser) {
  const ctx = await context(browser, { viewport: { width: 360, height: 780 }, isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  await page.goto(base + '/', { waitUntil: 'networkidle' });
  await page.keyboard.press('/');
  check(await page.evaluate(() => document.activeElement && document.activeElement.id === 'ask-q'), 'phone: "/" does not focus the Ask box');
  check((await page.textContent('#ask-help')).includes('no AI model'), 'phone: the Ask box does not say it holds no AI model');
  await page.click('.chips .chip >> nth=0');
  await page.waitForSelector('#answer .answer .open', { timeout: 5000 });
  await page.screenshot({ path: path.join(out, 'phone360-light-ask-chip.png'), fullPage: true });
  for (const [question, expected] of [['allocation by currency', 'By currency'], ['why is xtb failing', 'xtb-main'], ['buy more bonds', "can't answer that yet"]]) {
    await page.fill('#ask-q', question);
    await page.click('.ask-form button[type="submit"]');
    const answered = await page.waitForFunction(
      (want) => document.querySelector('#answer .answer') && document.querySelector('#answer').textContent.includes(want), expected, { timeout: 5000 },
    ).then(() => true, () => false);
    check(answered, `phone: the Ask box does not answer "${question}" with "${expected}"`);
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    check(overflow <= 0, `phone: the answer to "${question}" scrolls ${overflow} px sideways`);
    await page.screenshot({ path: path.join(out, `phone360-light-ask-${question.replace(/\W+/g, '-')}.png`), fullPage: true });
  }
  check(new URL(page.url()).search === '', 'phone: an answer in place changed the address');
  await page.click('#over-time nav a:has-text("Performance")');
  await page.waitForSelector('#over-time #history-title', { timeout: 5000 });
  await page.goto(base + '/inbox', { waitUntil: 'networkidle' });
  check(await page.isVisible('.ask-about .chip'), 'phone: the Inbox shows no Ask suggestions');
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  check(overflow <= 0, `phone: the Inbox with its Ask column scrolls ${overflow} px sideways`);
  await page.screenshot({ path: path.join(out, 'phone360-light-inbox-ask.png'), fullPage: true });
  await ctx.close();
}

// sideways returns how many pixels the page scrolls sideways.
async function sideways(page) {
  return page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
}

// recordPages checks the pages of D-195 at 360 px: an account's six tabs wrap without scrolling sideways and switch in
// place with htmx, the address following; an instrument's page shows its name, price chart, holdings, transactions,
// price checks and tax class, and the owner's rename form saves; the overview's chart card switches to the allocation,
// and a currency's bar opens the holdings behind it.
async function recordPages(browser) {
  const ctx = await context(browser, { viewport: { width: 360, height: 780 }, isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  await page.goto(base + '/accounts/acc-synthetic-1', { waitUntil: 'networkidle' });
  check(await page.locator('.tab-list a').count() === 6, 'phone: the account has not six tabs');
  check(await page.isVisible('nav.breadcrumb'), 'phone: the account has no breadcrumb');
  check(await sideways(page) <= 0, 'phone: the account with its tabs scrolls sideways');
  for (const [label, heading] of [['Lots', '#lots-title'], ['Reconciliation', '#checks-title'], ['Account settings', '#settings-title']]) {
    await page.click(`.tab-list a:has-text("${label}")`);
    const swapped = await page.waitForSelector(`#account-tabs ${heading}`, { timeout: 5000 }).then(() => true, () => false);
    check(swapped, `phone: the ${label} tab does not swap in`);
    check(new URL(page.url()).searchParams.has('tab'), `phone: the ${label} tab leaves the address without its tab`);
    check(await sideways(page) <= 0, `phone: the ${label} tab scrolls sideways`);
  }
  await page.screenshot({ path: path.join(out, 'phone360-light-account-settings-tab.png'), fullPage: true });
  await page.goto(base + '/accounts/acc-synthetic-1', { waitUntil: 'networkidle' });
  await page.click('#holdings-title ~ .table a[href^="/instruments/"] >> nth=0');
  await page.waitForSelector('#prices-title', { timeout: 5000 });
  for (const part of ['#prices-title', '#instrument-holdings-title', '#instrument-tx-title', '#instrument-checks-title', '#classes-title', '#rename-title']) {
    check(await page.locator(part).count() === 1, `phone: the instrument page has no ${part}`);
  }
  check(await page.locator('#prices-title ~ .chart canvas').count() > 0, 'phone: the instrument page draws no price chart');
  check(await sideways(page) <= 0, 'phone: the instrument page scrolls sideways');
  await page.screenshot({ path: path.join(out, 'phone360-light-instrument.png'), fullPage: true });
  await page.fill('#instrument-name', 'Synthetic World Fund');
  await Promise.all([page.waitForURL('**/instruments/XX0000000001?renamed=1'), page.click('form[action$="/name"] button[type="submit"]')]);
  check((await page.textContent('main')).includes('The name is saved.'), 'phone: the rename does not say it is saved');
  await page.screenshot({ path: path.join(out, 'phone360-light-instrument-renamed.png'), fullPage: true });
  await page.goto(base + '/', { waitUntil: 'networkidle' });
  await page.click('#over-time nav a:has-text("Allocation")');
  await page.waitForSelector('#over-time #allocation .bars', { timeout: 5000 });
  check(await sideways(page) <= 0, 'phone: the allocation tab scrolls sideways');
  await page.screenshot({ path: path.join(out, 'phone360-light-overview-allocation.png'), fullPage: true });
  await page.click('#allocation nav[aria-label="Split by"] a:has-text("Currency")');
  await page.waitForSelector('#allocation a.bar[href^="/holdings?by=currency"]', { timeout: 5000 });
  await Promise.all([page.waitForURL('**/holdings?**'), page.click('#allocation a.bar >> nth=0')]);
  check(await page.isVisible('#filtered-holdings-title'), 'phone: a currency bar does not open its holdings');
  check(await sideways(page) <= 0, 'phone: the holdings behind a bar scroll sideways');
  await ctx.close();
}

// app checks installability, the worker's cache, the offline page and sign-out, in a profile of its own: an
// incognito context is never installable.
// passkeyCheck creates a script token while the fake core wants a recent passkey check (D-183): the settings ask to
// confirm with a passkey, the button runs an assertion with Chromium's virtual authenticator, which holds a synthetic
// passkey for localhost, and the token is then created and its secret shown.
async function passkeyCheck(browser) {
  const ctx = await context(browser, { viewport: { width: 1280, height: 800 }, colorScheme: 'light' }, true);
  const page = await ctx.newPage();
  const cdp = await ctx.newCDPSession(page);
  await cdp.send('WebAuthn.enable');
  const { authenticatorId } = await cdp.send('WebAuthn.addVirtualAuthenticator', {
    options: { protocol: 'ctap2', transport: 'internal', hasResidentKey: true, hasUserVerification: true, isUserVerified: true, automaticPresenceSimulation: true },
  });
  const { privateKey } = generateKeyPairSync('ec', { namedCurve: 'P-256' });
  await cdp.send('WebAuthn.addCredential', {
    authenticatorId,
    credential: {
      credentialId: randomBytes(16).toString('base64'), isResidentCredential: true, rpId: 'localhost',
      privateKey: privateKey.export({ type: 'pkcs8', format: 'der' }).toString('base64'),
      userHandle: randomBytes(32).toString('base64'), signCount: 0,
    },
  });
  await page.goto(base + '/settings', { waitUntil: 'networkidle' });
  await page.fill('#token-name', 'nightly');
  await Promise.all([page.waitForLoadState('networkidle'), page.click('form[action="/settings/tokens"] button[type="submit"]')]);
  const confirm = page.locator('[data-webauthn="check"]');
  check(await confirm.isVisible(), 'passkey check: no "Confirm with your passkey" after a refused action');
  await page.screenshot({ path: path.join(out, 'desktop-light-passkey-check.png'), fullPage: true });
  // The page is already at /settings/tokens, the refused post's answer, so wait for the post again, not the URL.
  const posted = page.waitForResponse((r) => r.url() === base + '/settings/tokens' && r.request().method() === 'POST', { timeout: 30_000 });
  await confirm.click();
  await posted;
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: path.join(out, 'desktop-light-passkey-checked.png'), fullPage: true });
  const body = await page.textContent('body');
  check(body.includes('Script token nightly created'), 'passkey check: the action did not go ahead after the check');
  const { credentials } = await cdp.send('WebAuthn.getCredentials', { authenticatorId });
  check(credentials.length === 1 && credentials[0].signCount > 0, 'passkey check: the virtual authenticator signed nothing');
  await ctx.close();
}

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

// The IndexedDB database in which offline.js keeps the last overview (internal/views/ui/offline.go's offlineDB).
const offlineDB = 'openfief-offline';

// stored reads what the page's origin keeps besides the static cache: the IndexedDB databases' names, the kept
// overview's entries with its key's properties, and whether its bytes hold anything of the overview in plain text.
async function stored(page) {
  return page.evaluate(async (name) => {
    const dbs = (await indexedDB.databases()).map((d) => d.name);
    const out = { dbs, local: Object.keys(localStorage), session: Object.keys(sessionStorage) };
    if (!dbs.includes(name)) {
      return out;
    }
    const db = await new Promise((resolve, reject) => {
      const req = indexedDB.open(name);
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });
    const get = (what) => new Promise((resolve, reject) => {
      const req = what(db.transaction('overview', 'readonly').objectStore('overview'));
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });
    const entries = await get((s) => s.getAllKeys());
    const key = await get((s) => s.get('key'));
    const copy = await get((s) => s.get('copy'));
    db.close();
    out.entries = entries.map(String).sort();
    out.key = key && { cryptoKey: key instanceof CryptoKey, extractable: key.extractable, algorithm: key.algorithm.name, usages: key.usages.slice().sort() };
    out.exportRefused = await crypto.subtle.exportKey('raw', key).then(() => false, () => true);
    // Every way the overview could sit in the copy's bytes as text: UTF-8 read as Latin-1 keeps ASCII intact.
    const bytes = new Uint8Array(copy.data);
    const text = Array.from(bytes, (b) => String.fromCharCode(b)).join('');
    out.plain = ['Synthetic', 'PLN', 'sessionEnds', 'listed', 'core'].filter((s) => text.includes(s));
    out.sealed = copy.data instanceof ArrayBuffer && copy.iv instanceof Uint8Array && copy.iv.length === 12;
    return out;
  }, offlineDB);
}

// keptOverview checks D-151 in Chromium, in a profile of its own: opening the overview keeps it in IndexedDB,
// encrypted with AES-GCM under a key that can't be exported, and nothing else; offline, the overview's address shows
// that copy marked as offline with its time, and every other page the plain offline page; the copy is wiped on
// sign-out, when its session has ended, and when the server rejects the session.
async function keptOverview() {
  const profile = mkdtempSync(path.join(os.tmpdir(), 'openfief-ui-'));
  const ctx = await chromium.launchPersistentContext(profile, { ...browserOptions(), viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });
  const page = ctx.pages()[0] || await ctx.newPage();
  const signIn = async () => {
    await page.goto(base + '/test/sign-in', { waitUntil: 'networkidle' });
    await page.goto(base + '/', { waitUntil: 'networkidle' });
    await page.evaluate(() => navigator.serviceWorker.ready);
    await page.waitForFunction(async (name) => (await indexedDB.databases()).some((d) => d.name === name), offlineDB);
  };
  await signIn();
  // A switch of the allocation pushes its URL into the history, for which htmx would keep the page it leaves.
  const switches = page.locator('#allocation nav a:not([aria-current])');
  if (await switches.count() > 0) {
    await switches.first().click();
    await page.waitForLoadState('networkidle');
  }
  const kept = await stored(page);
  check(kept.dbs.length === 1 && kept.dbs[0] === offlineDB, `databases kept: ${JSON.stringify(kept.dbs)}`);
  // htmx remembers only the current path in sessionStorage, never a page.
  check(kept.local.length === 0 && kept.session.every((k) => k === 'htmx-current-path-for-history'),
    `kept in localStorage or sessionStorage: ${JSON.stringify([kept.local, kept.session])}`);
  check(JSON.stringify(kept.entries) === '["copy","key"]', `kept entries: ${JSON.stringify(kept.entries)}`);
  check(kept.key && kept.key.cryptoKey && kept.key.extractable === false && kept.key.algorithm === 'AES-GCM',
    `the key is not a non-extractable AES-GCM CryptoKey: ${JSON.stringify(kept.key)}`);
  check(kept.exportRefused, 'the key can be exported');
  check(kept.sealed, 'the copy is not ciphertext with a 12-byte IV');
  check(kept.plain && kept.plain.length === 0, `the kept copy holds plain text: ${JSON.stringify(kept.plain)}`);

  await ctx.setOffline(true);
  await page.goto(base + '/', { waitUntil: 'load' }).catch(() => {});
  const shown = await page.waitForSelector('#offline-copy:not([hidden])', { timeout: 5000 }).then(() => true, () => false);
  check(shown, 'offline: the overview shows no kept copy');
  const copy = shown ? await page.textContent('#offline-copy') : '';
  check(copy.includes('Offline copy.') && copy.includes('when you last opened it, at '), 'offline: the copy is not marked as offline with its time');
  check(copy.includes('Synthetic brokerage') && copy.includes('66.7%') && copy.includes('1 500,00 PLN'), 'offline: the copy lacks the account, a share or the total');
  const labelled = await page.locator('#offline-copy .table.stack td[data-label]').count();
  check(labelled > 0, 'offline: the copy\'s tables are not labelled for the phone');
  await page.screenshot({ path: path.join(out, 'phone-light-offline-overview.png'), fullPage: true });
  await page.goto(base + '/accounts/acc-synthetic-1', { waitUntil: 'load' }).catch(() => {});
  check(await page.isHidden('#offline-copy') && !(await page.textContent('body')).includes('Synthetic brokerage'),
    'offline: another page shows the kept overview');
  await ctx.setOffline(false);

  await page.goto(base + '/', { waitUntil: 'networkidle' });
  await Promise.all([page.waitForURL('**/login'), page.click('.menu-toggle').then(() => page.click('form[action="/logout"] button'))]);
  await page.waitForLoadState('networkidle');
  const out1 = await stored(page);
  check(!out1.dbs.includes(offlineDB), `sign-out left the kept overview: ${JSON.stringify(out1.dbs)}`);

  // The session ends (the fakes' sessions end on 2030-01-02): offline, the copy is wiped unread.
  await signIn();
  await ctx.clock.setFixedTime(new Date('2030-01-03T00:00:00Z'));
  await ctx.setOffline(true);
  await page.goto(base + '/', { waitUntil: 'load' }).catch(() => {});
  await page.waitForFunction(async (name) => !(await indexedDB.databases()).some((d) => d.name === name), offlineDB, { timeout: 5000 })
    .catch(() => {});
  check(await page.isHidden('#offline-copy'), 'offline: a copy whose session ended is shown');
  check(!(await stored(page)).dbs.includes(offlineDB), 'offline: a copy whose session ended is not wiped');
  await ctx.setOffline(false);
  await ctx.clock.setFixedTime(new Date());

  // The server rejects the session (expired or revoked at the core): the next page wipes the copy.
  await signIn();
  await page.goto(base + '/test/ended-session', { waitUntil: 'networkidle' });
  check(page.url().endsWith('/login'), `a rejected session opened ${page.url()}`);
  check(!(await stored(page)).dbs.includes(offlineDB), 'a rejected session left the kept overview');
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
  await removedRedirect(browser);
  await askBox(browser);
  await recordPages(browser);
  await passkeyCheck(browser);
  await app();
  await keptOverview();
} finally {
  await browser.close();
  process.kill(-server.pid);
}
if (failures.length > 0) {
  console.error(`${failures.length} browser check(s) failed`);
  process.exit(1);
}
console.log(`browser checks passed; screenshots in ${out}`);
