#!/usr/bin/env node
// Browser checks of the built-in UI over its in-memory fakes (task ui:browser). It starts them itself with
// TestServeForBrowser (go test) on ADDR, default localhost:8099, and stops them when done: a screenshot of every main page at 390 px in light and dark and at desktop width; no page
// scrolls sideways at 360 or 390 px; the menu folds the navigation away on a phone, with the Inbox's open count; the check
// pages D-192 removed land on the Inbox at 360 px with its in-place forms, and an account's banner links there; the Inbox shows the
// maintainer's proposal with its figures, Apply and Dismiss, and the model portfolio's page its latest note (D-204); the manifest makes the UI
// installable; the service worker caches only static files, serves the offline page instead of a page when the
// network is gone, and empties its cache on sign-out; an action the fake core refuses for want of a recent passkey
// check (D-183) asks to confirm with a passkey, runs the check with Chromium's virtual authenticator, and goes ahead;
// at 360 px the overview's Ask box (D-194) takes "/", answers a suggestion chip, a typed command and one it doesn't
// understand, its chart card switches to the performance, and the Inbox shows its side column of Ask suggestions; with a
// model (D-203) it says where a question goes, shows the model's text with the card it read, a refusal with the box's commands as chips, a clarifying answer with its choices, and refuses a question no model may receive;
// at 360 px an account's tabs wrap and switch in place (D-195), an instrument's page shows its parts and takes the
// owner's rename, and the overview's chart card switches to the allocation, whose currency bar opens its holdings; the overview is kept offline only
// encrypted, under a key that can't be exported, shown offline marked as a copy, and wiped on sign-out and when the
// session ends (D-151); the Access page's People section invites a viewer (one-time link) and asks to confirm a revoke, and a viewer
// granted nothing sees "Nothing has been shared with you yet" on every page (D-210); Settings' Audit log filters by its form and, while
// changes' entries wait to enter the log, shows the audit-log-behind item's words in a banner (D-211).
// The Ask box's progress and layout (D-194, D-203): a question in flight shows itself under the field with a quiet indicator (still under
// reduced motion), the Ask button off and "Thinking…" or "Looking it up…"; the answer sits directly under the field and in view with one
// attribution line, the help is folded behind "How this works", and the suggestions are one row that scrolls on its own at 360 px.
// With Cloudflare Turnstile in front of the Ask box's model (D-208), on a second UI over the fakes whose Cloudflare is a fake too: the check
// runs on demand. The policy allows Cloudflare's script and frame and nothing else outside; a command is answered with no check; a model
// question goes without a token, is told a check is needed, runs the widget (explicit render, execute, interaction-only) and goes again with
// its one-time token, the same question, and is answered; nothing is checked again afterwards; an interaction-needed widget gets time;
// an error, a timeout, a blocked script and an unsupported browser each give one message with the code; a stalled widget becomes a visible
// one; the page reports each state and the console holds no policy violation; the first UI has no widget.
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
// A second UI over the fakes has Cloudflare Turnstile in front of the Ask box's model (D-208), checked against a fake Cloudflare.
const turnstileAddr = process.env.ADDR_TURNSTILE || 'localhost:8098';
const turnstileBase = `http://${turnstileAddr}`;

// serve starts the UI over its fakes on at, with the environment extra, in a process group of its own and waits until it answers.
async function serve(at = addr, extra = {}) {
  const server = spawn('go', ['test', './internal/views/ui/', '-run', '^TestServeForBrowser$', '-count=1', '-timeout', '15m'], {
    env: { ...process.env, OPENFIEF_UI_SERVE: at, ...extra }, stdio: ['ignore', 'inherit', 'inherit'], detached: true,
  });
  for (let i = 0; i < 120; i++) {
    try {
      if ((await fetch(`http://${at}/healthz`)).ok) {
        return server;
      }
    } catch {
      // not listening yet
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
  process.kill(-server.pid);
  throw new Error(`the UI over its fakes did not answer on http://${at}`);
}

// The fakes' one account is acc-synthetic-1 (internal/views/ui/fakes_test.go).
const pages = {
  overview: '/', 'overview-ask': '/?q=what+needs+me', 'overview-performance': '/?over=performance', 'overview-allocation': '/?over=allocation',
  accounts: '/accounts', account: '/accounts/acc-synthetic-1', 'account-transactions': '/accounts/acc-synthetic-1?tab=transactions',
  'account-lots': '/accounts/acc-synthetic-1?tab=lots', 'account-cash': '/accounts/acc-synthetic-1?tab=cash',
  'account-reconciliation': '/accounts/acc-synthetic-1?tab=reconciliation', 'account-settings': '/accounts/acc-synthetic-1?tab=settings',
  instrument: '/instruments/XX0000000001?account=acc-synthetic-1', 'instrument-isin-pending': '/instruments/PLCOI1130004', holdings: '/holdings?by=currency&key=EUR', inbox: '/inbox',
  'inbox-acknowledged': '/inbox?view=acknowledged', 'inbox-done': '/inbox?view=done', sleeves: '/sleeves', plugins: '/plugins',
  settings: '/settings', models: '/settings/models', access: '/access', audit: '/settings/audit', 'audit-filtered': '/settings/audit?outcome=refused&from=2026-03-01&to=2026-03-03', setup: '/setup', model: '/model',
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
  for (const form of ['/inbox/tickers/confirm', '/inbox/retry-sync', '/accounts/bind', '/inbox/price-checks/acknowledge', '/inbox/proposals/apply', '/inbox/proposals/dismiss']) {
    check(await page.locator(`form[action="${form}"]`).count() > 0, `phone: the Inbox has no ${form} form`);
  }
  // The maintainer's proposal card (D-204): the figures behind the change are in the page and the table fits the phone.
  await page.goto(base + '/inbox', { waitUntil: 'networkidle' });
  check((await page.textContent('#item-70')).includes('63.1%'), "phone: the proposal's card shows no current weight");
  check(await page.locator('#item-70 form[action="/inbox/proposals/apply"] button').isVisible(), 'phone: the proposal has no visible Apply button');
  const proposalOverflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  check(proposalOverflow <= 0, `phone: the proposal's card scrolls ${proposalOverflow} px sideways`);
  await page.screenshot({ path: path.join(out, 'phone360-light-inbox-proposal.png'), fullPage: true });
  // The model portfolio's page: the maintainer's latest note and the owner's Ask button.
  await page.goto(base + '/model', { waitUntil: 'networkidle' });
  check(await page.isVisible('#maintainer-note'), "phone: the model portfolio's page shows no maintainer note");
  check(await page.locator('form[action="/model/maintainer/ask"] button').isVisible(), 'phone: the owner has no Ask the maintainer button');
  const modelOverflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  check(modelOverflow <= 0, `phone: the model portfolio's page scrolls ${modelOverflow} px sideways`);
  await page.screenshot({ path: path.join(out, 'phone360-light-model.png'), fullPage: true });
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
  for (const [question, expected] of [['allocation by currency', 'By currency'], ['why is xtb failing', 'the latest with the error class timeout'], ['buy more bonds', "can't answer that yet"]]) {
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

// askModelBox checks the Ask box with a model (D-203) at phone width, in light and dark: the box says a question beyond its
// commands goes to the model; the answer shows the model's text and the card it read; a refusal shows the box's commands as
// chips and a clarifying answer its candidates as choices, each of which the box answers itself at once; a model that may not
// receive the question's data tier (T0) is refused with nothing sent, and the page says so; a command still answers as before;
// and none of it scrolls sideways. The model is the fakes' (browserAskModel); it is taken away again at the end.
async function askModelBox(browser) {
  for (const scheme of ['light', 'dark']) {
    const ctx = await context(browser, { viewport: { width: 360, height: 780 }, colorScheme: scheme, isMobile: true, hasTouch: true }, true);
    const page = await ctx.newPage();
    const ask = async (question, selector) => {
      await page.fill('#ask-q', question);
      await page.click('.ask-form button[type="submit"]');
      await page.waitForSelector(selector, { timeout: 5000 });
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
      check(overflow <= 0, `phone ${scheme}: the answer to "${question}" scrolls ${overflow} px sideways`);
    };
    await page.goto(base + '/test/ask-model?tiers=T0,T2', { waitUntil: 'networkidle' });
    check((await page.textContent('#ask-help')).includes('goes to the model'), `phone ${scheme}: the Ask box does not say where a question goes`);
    await ask('how diversified am I', '#answer .used-card');
    const text = await page.textContent('#answer .model-text');
    check(text.includes('Most of it is listed.') && text.includes('The rest is cash.'), `phone ${scheme}: the model's text is not shown`);
    check((await page.textContent('#answer .used-card')).includes('By asset class'), `phone ${scheme}: the card the model read is not shown`);
    // D-222: a model's answer says who answered, that it can be wrong, and links to the eval's results.
    const note = await page.textContent('#answer .ask-note');
    check(note.includes('Answered by') && note.includes('used 1 card') && ((await page.textContent('#answer')).match(/Answered by/g) || []).length === 1 && note.includes('AI can be wrong; check the cards. Not investment advice.') && note.includes('How models did in our evaluation'),
      `phone ${scheme}: a model's answer does not show its name, the disclaimer and the link to the evaluation`);
    const evalHref = await page.getAttribute('#answer .ask-note a.ask-eval-link', 'href');
    check(evalHref === 'https://openfief.dev/docs/reference/ask-eval/', `phone ${scheme}: the evaluation link is ${evalHref}`);
    await page.screenshot({ path: path.join(out, `phone360-${scheme}-ask-model.png`), fullPage: true });
    await ask('allocation by currency', '#answer:has-text("By currency")');
    check((await page.locator('#answer .ask-note').count()) === 0, `phone ${scheme}: a command's answer shows the model note`);
    check((await page.textContent('#answer')).includes('no AI model') && (await page.locator('#answer .used-card').count()) === 0,
      `phone ${scheme}: a command is not answered by the box itself`);
    // A refusal: the model's words, and the box's own four commands as chips (D-194).
    await ask('sell everything', '#answer .ask-suggestions .chip');
    const refusal = await page.textContent('#answer');
    check(refusal.includes("I can only read, so I can't sell anything.") && refusal.includes('From ') && !refusal.includes("it can't do that") && (await page.locator('#answer .ask-suggestions .chip').count()) === 4
      && (await page.locator('#answer .ask-choices').count()) === 0 && (await page.locator('#answer .used-card').count()) === 0,
      `phone ${scheme}: a refusal does not show the model's words and the box's four command chips`);
    check((await page.locator('#answer .ask-suggestions .chip').first().boundingBox()).height >= 44, `phone ${scheme}: a command chip is under 44 px tall`);
    await page.screenshot({ path: path.join(out, `phone360-${scheme}-ask-model-refuse.png`), fullPage: true });
    await page.click('#answer .ask-suggestions .chip:has-text("Allocation by currency")');
    await page.waitForSelector('#answer:has-text("no AI model")', { timeout: 5000 });
    check((await page.textContent('#answer')).includes('By currency'), `phone ${scheme}: a command chip of a refusal is not answered by the box itself`);
    // A clarifying answer: the model's question, and its candidates as choices that ask the box's own show command.
    await ask('i mean the fund', '#answer .ask-choices .chip');
    const clarify = await page.textContent('#answer');
    check(clarify.includes('Which one do you mean?') && !clarify.includes('it asks which you mean') && (await page.locator('#answer .ask-choices .chip').count()) === 2
      && (await page.locator('#answer .ask-suggestions').count()) === 0,
      `phone ${scheme}: a clarifying answer does not show its question and two choices`);
    await page.screenshot({ path: path.join(out, `phone360-${scheme}-ask-model-clarify.png`), fullPage: true });
    await page.click('#answer .ask-choices .chip:has-text("Synthetic Fund")');
    await page.waitForSelector('#answer:has-text("no AI model")', { timeout: 5000 });
    check((await page.textContent('#answer')).includes('Synthetic Fund') && (await page.locator('#answer .ask-choices').count()) === 0,
      `phone ${scheme}: a choice of a clarifying answer is not answered by the box itself`);
    await page.goto(base + '/test/ask-model?tiers=T2', { waitUntil: 'networkidle' });
    await ask('how diversified am I', '#answer .answer-title');
    check((await page.textContent('#answer .answer-title')).includes('No model may see this question') && (await page.textContent('#answer')).includes('Nothing was sent')
      && (await page.locator('#answer .used-card').count()) === 0,
      `phone ${scheme}: the refusal of a question no model may receive is not shown, or a card is drawn anyway`);
    await page.screenshot({ path: path.join(out, `phone360-${scheme}-ask-model-refused.png`), fullPage: true });
    await page.goto(base + '/test/ask-model?tiers=off', { waitUntil: 'networkidle' });
    check((await page.textContent('#ask-help')).includes('no AI model'), `phone ${scheme}: the Ask box does not go back to no model`);
    await ctx.close();
  }
}

// The widget's script, as Cloudflare would serve it, for the page to load where the check's script is allowed. It models the part of
// Turnstile's explicit API the page uses: render(container, options) returns a widget id and runs nothing (the page renders with
// execution "execute"), execute(id) starts the challenge, reset(id) and remove(id) clear or drop the widget. A token is "ok-browser-N",
// which the fake Cloudflare accepts once, put in a hidden field inside the container as the real widget does and passed to the
// options' callback. window.__fakeTurnstileMode, set before the page's scripts run, says what execute() does: "auto" makes a
// token at once, "manual" does nothing until the test calls succeed(), and "click" first calls the before-interactive callback,
// as Cloudflare does when it needs a click. The test reads what the page did through window.__fakeTurnstile (renders, executes,
// resets, removes) and plays Cloudflare through succeed(), fail(code), expire() and unsupported(), the last three returning what
// the page's callback returned.
const fakeTurnstileScript = `(function () {
  var mode = window.__fakeTurnstileMode || 'auto';
  var n = window.__fakeTurnstileSeed || 0; // a token is good once, so each page's tokens start where no other page's did
  var widgets = [];
  var renders = []; // what turnstile.render was called with, for the test to read
  var executes = 0, resets = 0, removes = 0;
  function current() { return widgets[widgets.length - 1]; }
  function call(name, argument) {
    var w = current();
    var given = w && w.options[name];
    return typeof given === 'function' ? given(argument) : undefined;
  }
  function succeed() {
    var w = current();
    if (!w) { return; }
    var input = w.div.querySelector('input[name="cf-turnstile-response"]');
    if (!input) {
      input = document.createElement('input');
      input.type = 'hidden';
      input.name = 'cf-turnstile-response';
      w.div.appendChild(input);
    }
    n += 1;
    input.value = 'ok-browser-' + n;
    call('callback', input.value);
  }
  window.turnstile = {
    render: function (div, options) {
      renders.push({ sitekey: options.sitekey, appearance: options.appearance, execution: options.execution, retry: options.retry,
        callbacks: Object.keys(options).filter(function (k) { return typeof options[k] === 'function'; }).sort() });
      widgets.push({ div: div, options: options });
      return 'fake-widget-' + renders.length;
    },
    execute: function () {
      executes += 1;
      if (mode === 'auto') { Promise.resolve().then(succeed); }
      if (mode === 'click') { call('before-interactive-callback'); }
    },
    reset: function () {
      resets += 1;
      var w = current();
      var input = w && w.div.querySelector('input[name="cf-turnstile-response"]');
      if (input) { input.value = ''; }
    },
    remove: function () {
      removes += 1;
      var w = widgets.pop();
      var input = w && w.div.querySelector('input[name="cf-turnstile-response"]');
      if (input) { input.remove(); }
    }
  };
  window.__fakeTurnstile = {
    renders: function () { return renders; },
    executes: function () { return executes; },
    resets: function () { return resets; },
    removes: function () { return removes; },
    succeed: succeed,
    fail: function (code) { return call('error-callback', code); },
    expire: function () { return call('expired-callback'); },
    unsupported: function () { return call('unsupported-callback'); }
  };
})();`;

let nextSeed = 0; // the fake Cloudflare takes each token once, so every page after the first makes its own range
const checkingText = 'Checking you\'re a person…';

// turnstileContext opens a signed-in page on the Turnstile UI with the fake widget in the given mode (and a fake clock when asked), counting
// the requests the Ask box makes and noting whether the no-token card was ever drawn.
async function turnstileContext(browser, mode, options = {}) {
  const ctx = await browser.newContext({ viewport: { width: options.width || 360, height: 780 }, colorScheme: 'light', isMobile: true, hasTouch: true });
  const seed = mode === 'auto' ? 0 : (nextSeed += 100);
  await ctx.addInitScript(([m, sd]) => {
    window.__fakeTurnstileMode = m;
    window.__fakeTurnstileSeed = sd;
    new MutationObserver(() => {
      const a = document.getElementById('answer');
      if (a && a.textContent.includes('A check comes before the model answers')) { window.__sawNoTokenCard = true; }
    }).observe(document, { childList: true, subtree: true, characterData: true });
  }, [mode, seed]);
  if (options.block) {
    await ctx.route('https://challenges.cloudflare.com/**', (route) => route.abort());
  } else {
    await ctx.route('https://challenges.cloudflare.com/**', (route) => route.fulfill({ status: 200, contentType: 'text/javascript', body: fakeTurnstileScript }));
  }
  if (options.clock) {
    await ctx.clock.install();
  }
  const page = await ctx.newPage();
  const asks = []; // the query strings of the box's requests, in order
  const posts = []; // the bodies of the page's reports about itself (D-227)
  page.on('request', (r) => {
    if (r.url().includes('/ask?')) { asks.push(new URL(r.url()).searchParams); }
    if (r.method() === 'POST' && r.url().endsWith('/client-events')) { posts.push(r.postData() || ''); }
  });
  await page.goto(turnstileBase + '/test/sign-in');
  await page.goto(turnstileBase + '/test/ask-model?tiers=T0,T2', { waitUntil: 'networkidle' });
  const fake = (what) => page.evaluate((w) => (w === 'renders' ? window.__fakeTurnstile.renders() : window.__fakeTurnstile[w]()), what);
  const submit = async (q) => {
    await page.fill('#ask-q', q);
    await page.click('.ask-form button[type="submit"]');
  };
  return { ctx, page, asks, posts, fake, submit };
}

// turnstileBox checks the Ask box with Turnstile in front of its model (D-208), checked on demand: the page's policy allows Cloudflare's
// script and frame and no other outside host; the widget is not rendered, let alone executed, while the person only reads, types or asks
// a command; a question to the model goes without a token first, is told a check is needed (and the no-token card is never drawn), runs
// the check, and goes again with the token, the same question, and is answered; the token is not kept (the form never holds one) and
// nothing is checked again after the answer, only when the next model question needs it; and the console holds no policy violation.
async function turnstileBox(browser) {
  const { ctx, page, asks, fake, submit } = await turnstileContext(browser, 'auto');
  const violations = [];
  page.on('console', (message) => {
    if (message.text().includes('Content Security Policy')) { violations.push(message.text()); }
  });
  const response = await page.goto(turnstileBase + '/test/ask-model?tiers=T0,T2', { waitUntil: 'networkidle' });
  const policy = response.headers()['content-security-policy'] || '';
  check(policy.includes('script-src \'self\' https://challenges.cloudflare.com') && policy.includes('frame-src https://challenges.cloudflare.com')
    && (policy.match(/https:\/\//g) || []).length === 2 && !policy.includes('unsafe'), `turnstile: the page's policy is ${policy}`);
  check((await page.locator('.ask-box .ask-check[data-sitekey]').count()) === 1 && (await page.locator('.cf-turnstile').count()) === 0,
    'turnstile: the page has no container for an explicit widget, or it renders one by itself');
  check((await page.locator('#ask-check-status').count()) === 0, 'turnstile: the page still has a status line for the check');
  check((await page.locator('.ask-form input[name="cf-turnstile-response"]').count()) === 0, 'turnstile: the form holds a token field before any check');
  check((await page.getAttribute('script[src*="challenges.cloudflare.com"]', 'src')).endsWith('?render=explicit'), 'turnstile: the widget script is not loaded for explicit render');
  await page.waitForTimeout(1000);
  check((await fake('renders')).length === 0 && (await fake('executes')) === 0, 'turnstile: the widget was rendered or run before any question needed it');
  // A command is answered at once, with no check.
  await submit('what needs me');
  await page.waitForSelector('#answer:has-text("no AI model")', { timeout: 5000 });
  check(asks.length === 1 && !asks[0].has('cf-turnstile-response') && (await fake('executes')) === 0, `turnstile: a command asked ${asks.length} times or ran the check`);
  // A model question: no token first, then the check, then the same question with its token.
  const question = 'how diversified am I';
  await submit(question);
  await page.waitForSelector('#answer .used-card', { timeout: 5000 });
  check(asks.length === 3 && !asks[1].has('cf-turnstile-response') && asks[1].get('q') === question, `turnstile: the question did not go first without a token (${asks.length} requests)`);
  check(asks[2].get('q') === question && asks[2].get('cf-turnstile-response') === 'ok-browser-1' && asks[2].get('ask-check-path') === 'invisible',
    `turnstile: the question did not go again with its token: ${asks[2]}`);
  const rendered = await fake('renders');
  check(rendered.length === 1 && rendered[0].appearance === 'interaction-only' && rendered[0].execution === 'execute' && rendered[0].sitekey.length > 0
    && rendered[0].callbacks.join() === 'before-interactive-callback,callback,error-callback,expired-callback,unsupported-callback',
  `turnstile: the widget was rendered as ${JSON.stringify(rendered)}`);
  check((await fake('executes')) === 1, 'turnstile: the check did not run exactly once for the question');
  check(!(await page.evaluate(() => window.__sawNoTokenCard === true)), 'turnstile: the no-token card was drawn on the way');
  const text = await page.textContent('#answer');
  check((text.match(/Answered by/g) || []).length === 1 && text.includes('AI can be wrong; check the cards. Not investment advice.'),
    'turnstile: the answer does not say who answered exactly once');
  check((await page.locator('.ask-form input[name="cf-turnstile-response"]').count()) === 0, 'turnstile: the form holds a token after the check');
  check(await page.isHidden('#ask-pending') && await page.isEnabled('.ask-form button[type="submit"]'), 'turnstile: after the answer the progress shows or the button is disabled');
  // Nothing is checked again by itself: no reset, no second execute, however long the person reads.
  await page.waitForTimeout(2000);
  check((await fake('executes')) === 1 && (await fake('resets')) === 0 && (await fake('renders')).length === 1, 'turnstile: the check ran again after the answer');
  await page.screenshot({ path: path.join(out, 'phone360-light-ask-turnstile.png'), fullPage: true });
  // The next model question needs a new check: the widget is reset and run again, once.
  await submit('how diversified am I really');
  await page.waitForSelector('#answer .bubble:has-text("how diversified am I really")', { timeout: 5000 });
  await page.waitForSelector('#answer .used-card', { timeout: 5000 });
  check(asks.length === 5 && asks[4].get('cf-turnstile-response') === 'ok-browser-2' && (await fake('executes')) === 2 && (await fake('resets')) === 1,
    `turnstile: the second question did not make one new check (${asks.length} requests, ${await fake('executes')} runs)`);
  const overflow = await sideways(page);
  check(overflow <= 0, `turnstile: the answer scrolls ${overflow} px sideways at 360 px`);
  check(violations.length === 0, `turnstile: the console holds policy violations: ${violations.join('; ')}`);
  await ctx.close();
}

// turnstileOnDemand checks the states of the check against a widget the test drives (the fake's "manual" mode, a fake clock for the
// waits): while it runs the box shows the question and "Checking you're a person…" with the Ask button off; the question that goes
// again is the one asked, not what the field says by then; an error, a timeout, a blocked script and an unsupported browser each give ONE
// message in the answer with the code (and Brave's hint where it helps), after which the box asks again with a new check; a widget that
// asks for a click gets time; a check that stalls is swapped for a visible one that later checks start with; and the page reports each
// state (D-227) without a word typed or a token.
async function turnstileOnDemand(browser) {
  // The check waits for Cloudflare; the question that goes again is the one that was asked.
  {
    const { ctx, page, asks, fake, submit } = await turnstileContext(browser, 'manual');
    await submit('how diversified am I');
    await page.waitForSelector('#ask-pending.is-checking', { timeout: 5000 });
    check((await page.textContent('#ask-pending-label')) === checkingText && (await page.textContent('#ask-pending-q')) === 'how diversified am I',
      `turnstile on demand: while the check runs the box says ${await page.textContent('#ask-pending-label')}`);
    check(await page.isDisabled('.ask-form button[type="submit"]') && await page.isHidden('#answer'), 'turnstile on demand: the button is on, or the old answer shows, during the check');
    check((await page.locator('#ask-check-status').count()) === 0 && (await fake('executes')) === 1, 'turnstile on demand: a status line exists or the check did not start once');
    await page.screenshot({ path: path.join(out, 'phone360-light-ask-turnstile-checking.png'), fullPage: true });
    await page.fill('#ask-q', 'something else entirely');
    await fake('succeed');
    await page.waitForSelector('#answer .used-card', { timeout: 5000 });
    check(asks.length === 2 && asks[1].get('q') === 'how diversified am I' && asks[1].get('cf-turnstile-response') === 'ok-browser-101',
      `turnstile on demand: the question that went again was ${asks[asks.length - 1]}`);
    check(await page.isEnabled('.ask-form button[type="submit"]') && await page.isHidden('#ask-pending'), 'turnstile on demand: the box is still busy after the answer');
    // An error: one message with the code and the hint; the box is free; the next question makes a new check.
    await page.fill('#ask-q', 'and by currency');
    await page.click('.ask-form button[type="submit"]');
    await page.waitForSelector('#ask-pending.is-checking', { timeout: 5000 });
    const handled = await page.evaluate(() => window.__fakeTurnstile.fail('110200'));
    check(handled === true, `turnstile on demand: the error callback returned ${handled}, not true`);
    await page.waitForSelector('#answer .answer-title:has-text("Couldn\'t confirm you\'re a person")', { timeout: 5000 });
    const message = await page.textContent('#answer');
    check(message.includes('Cloudflare error 110200') && message.includes('Brave') && (await page.locator('#answer .answer').count()) === 1 && message.includes('and by currency'),
      `turnstile on demand: the error says ${message}`);
    check(await page.isEnabled('.ask-form button[type="submit"]') && await page.isHidden('#ask-pending') && !(await page.evaluate(() => window.__sawNoTokenCard === true)),
      'turnstile on demand: after an error the box is busy or the no-token card was drawn');
    await page.screenshot({ path: path.join(out, 'phone360-light-ask-turnstile-error.png'), fullPage: true });
    await submit('and by currency');
    await page.waitForSelector('#ask-pending.is-checking', { timeout: 5000 });
    check((await fake('executes')) === 3 && (await fake('resets')) === 2, `turnstile on demand: asking again after an error ran the check ${await fake('executes')} times in all`);
    await page.evaluate(() => window.__fakeTurnstile.unsupported());
    await page.waitForSelector('#answer:has-text("unsupported")', { timeout: 5000 });
    check((await page.textContent('#answer')).includes('This browser can\'t run the check'), 'turnstile on demand: an unsupported browser is not told so');
    await ctx.close();
  }
  // Cloudflare asks for a click: the person gets time, and the widget is not swapped for another.
  {
    const { ctx, page, asks, fake, submit } = await turnstileContext(browser, 'click', { clock: true });
    await submit('how diversified am I');
    await present(page, '#ask-pending-label:has-text("Complete the check if it asks.")');
    await ctx.clock.fastForward(30000);
    check((await page.locator('#answer .answer-title:has-text("Couldn")').count()) === 0 && (await fake('renders')).length === 1 && (await fake('removes')) === 0,
      'turnstile on demand: a widget that asked for a click timed out, or was swapped, within 30 seconds');
    await page.evaluate(() => window.__fakeTurnstile.succeed());
    await settle(ctx);
    await present(page, '#answer .used-card');
    await ctx.close();
  }
  // A check that does not finish in 15 seconds is called timed out; before that, the stalled invisible one is swapped for a visible one.
  {
    const { ctx, page, asks, posts, fake, submit } = await turnstileContext(browser, 'manual', { clock: true });
    await submit('how diversified am I');
    await present(page, '#ask-pending.is-checking');
    await ctx.clock.fastForward(7000);
    check((await fake('renders')).length === 1 && (await page.textContent('#ask-pending-label')) === checkingText, 'turnstile on demand: the widget was swapped before it had waited');
    check((await page.inputValue('#ask-check-path')) === 'invisible', 'turnstile on demand: the path was not invisible before the fallback');
    await ctx.clock.fastForward(1500);
    const rendered = await fake('renders');
    check(rendered.length === 2 && rendered[1].appearance === 'always' && rendered[1].execution === 'execute' && (await fake('removes')) === 1 && (await fake('executes')) === 2,
      `turnstile on demand: the stalled widget became ${JSON.stringify(rendered)}`);
    check((await page.textContent('#ask-pending-label')) === checkingText + ' Complete the check below.' && (await page.inputValue('#ask-check-path')) === 'visible-fallback',
      `turnstile on demand: after the swap the box says ${await page.textContent('#ask-pending-label')}`);
    await page.screenshot({ path: path.join(out, 'phone360-light-ask-turnstile-fallback.png'), fullPage: true });
    await ctx.clock.fastForward(16000);
    await present(page, '#answer:has-text("The check timed out (timeout)")');
    check((await page.textContent('#answer')).includes('Brave') && await page.isEnabled('.ask-form button[type="submit"]'), 'turnstile on demand: the timeout message lacks the hint, or the box stays busy');
    // The next question starts with the visible widget, which the person ticks: the question goes with its path.
    await submit('how diversified am I');
    await present(page, '#ask-pending.is-checking');
    check((await fake('renders')).length === 2 && (await fake('removes')) === 1 && (await fake('executes')) === 3 && (await fake('resets')) === 1,
      'turnstile on demand: the next check did not start with the visible widget');
    await page.evaluate(() => window.__fakeTurnstile.succeed());
    await settle(ctx);
    await present(page, '#answer .used-card');
    const last = asks[asks.length - 1];
    check(last.get('ask-check-path') === 'visible-fallback' && last.get('cf-turnstile-response').startsWith('ok-browser-'), `turnstile on demand: the question went as ${last}`);
    // What the page reported (D-227), with no word typed and no token.
    await page.evaluate(() => window.__fakeTurnstile.fail('110200'));
    await ctx.clock.fastForward(3000);
    await new Promise((resolve) => setTimeout(resolve, 500)); // the beacons leave the page after the clock moves
    const sentEvents = posts.flatMap((p) => JSON.parse(p).events);
    const kinds = sentEvents.map((e) => e.k);
    check(['turnstile-execute', 'turnstile-script-loaded', 'turnstile-timeout', 'turnstile-fallback-shown', 'turnstile-error', 'turnstile-success'].every((k) => kinds.includes(k)),
      `client events: the page reported ${kinds.join(', ')}`);
    check(kinds.indexOf('turnstile-execute') < kinds.indexOf('turnstile-timeout') && kinds.indexOf('turnstile-timeout') < kinds.indexOf('turnstile-fallback-shown'),
      `client events: the states came in the order ${kinds.join(', ')}`);
    const success = sentEvents.find((e) => e.k === 'turnstile-success');
    check(success && success.path === 'visible-fallback' && Number.isInteger(success.ms), `client events: the success was ${JSON.stringify(success)}`);
    const timeout = sentEvents.find((e) => e.k === 'turnstile-error' && e.code === 'timeout') || sentEvents.find((e) => e.k === 'turnstile-error');
    check(timeout !== undefined, 'client events: the failure was not reported');
    const everything = posts.join('\n');
    check(!everything.includes('how diversified') && !/ok-browser-/.test(everything), 'client events: the payload holds what was typed or a token');
    const allowed = new Set(['k', 'ms', 'code', 'path', 'msg', 'src', 'line', 'status', 'brave']);
    check(sentEvents.every((e) => Object.keys(e).every((key) => allowed.has(key))), 'client events: an event carries a field outside the vocabulary');
    await ctx.close();
  }
  // Cloudflare's script is blocked (an extension, Brave's Shields): one message that says to allow it.
  {
    const { ctx, page, submit } = await turnstileContext(browser, 'manual', { clock: true, block: true });
    await submit('how diversified am I');
    await present(page, '#ask-pending.is-checking');
    await ctx.clock.fastForward(9000);
    await present(page, '#answer:has-text("Your browser blocked the check (blocked)")');
    const said = await page.textContent('#answer');
    check(said.includes('Allow challenges.cloudflare.com for this site (in Brave: turn Shields off for it), then reload.'), `turnstile on demand: with the script blocked the answer says ${said}`);
    await ctx.close();
  }
}

// askProgress checks, with no Turnstile and a model, what a person sees while a question is in flight (the response is held back): the
// question echoed under the field with a quiet animated indicator (a still one under prefers-reduced-motion), "Thinking…" for a model
// and "Looking it up…" for a chip, the Ask button off and the old answer out of the way; then the answer directly under the field and
// in view, one attribution line, the help folded away behind "How this works", the suggestions in one row that scrolls on its own and
// no sideways page scroll at 360 px.
async function askProgress(browser) {
  for (const reduced of [false, true]) {
    const ctx = await context(browser, { viewport: { width: 360, height: 780 }, isMobile: true, hasTouch: true, reducedMotion: reduced ? 'reduce' : 'no-preference' }, true);
    const page = await ctx.newPage();
    await page.goto(base + '/test/ask-model?tiers=T0,T2', { waitUntil: 'networkidle' });
    const tag = reduced ? 'reduced motion' : 'motion';
    check((await page.locator('details.ask-howto:not([open])').count()) === 1 && await page.isHidden('#ask-help'), `ask ux (${tag}): the help is not folded behind "How this works"`);
    check((await page.textContent('details.ask-howto summary')) === 'How this works' && (await page.locator('#ask-check-status').count()) === 0, `ask ux (${tag}): the summary or a status line is wrong`);
    check((await page.locator('.chips-row').count()) === 1 && (await page.locator('.chips-row .chip').count()) > 1, `ask ux (${tag}): the suggestions are not in one group`);
    const tops = await page.locator('.chips-row .chip').evaluateAll((els) => els.map((e) => Math.round(e.getBoundingClientRect().top)));
    const heights = await page.locator('.chips-row .chip').evaluateAll((els) => els.map((e) => e.getBoundingClientRect().height));
    check(new Set(tops).size === 1 && heights.every((h) => h >= 44), `ask ux (${tag}): the chips are not one row of tall targets: ${tops} ${heights}`);
    check((await page.locator('.chips-row').evaluate((el) => el.scrollWidth > el.clientWidth)) && (await sideways(page)) <= 0, `ask ux (${tag}): the chips row does not scroll on its own, or the page does`);
    await page.screenshot({ path: path.join(out, `phone360-light-ask-home-${reduced ? 'reduced' : 'motion'}.png`) });
    // The response is held back so the in-flight state can be seen.
    let release;
    const held = new Promise((resolve) => { release = resolve; });
    await page.route('**/ask?*', async (route) => { await held; await route.continue(); });
    await page.fill('#ask-q', 'how diversified am I');
    await page.click('.ask-form button[type="submit"]');
    await page.waitForSelector('#ask-pending.htmx-request', { state: 'visible', timeout: 5000 });
    check((await page.textContent('#ask-pending-q')) === 'how diversified am I' && (await page.textContent('#ask-pending-label')) === 'Thinking…',
      `ask ux (${tag}): the pending answer says ${await page.textContent('#ask-pending-q')} / ${await page.textContent('#ask-pending-label')}`);
    check(await page.isDisabled('.ask-form button[type="submit"]'), `ask ux (${tag}): the Ask button is on while a question is in flight`);
    const animation = await page.locator('.dots i').first().evaluate((el) => getComputedStyle(el).animationName);
    check(reduced ? animation === 'none' : animation !== 'none', `ask ux (${tag}): the indicator's animation is ${animation}`);
    const pendingBox = await page.locator('#ask-pending').boundingBox();
    const fieldBox = await page.locator('.ask-field').boundingBox();
    check(pendingBox.y > fieldBox.y && pendingBox.y - (fieldBox.y + fieldBox.height) < 120, `ask ux (${tag}): the pending answer is not directly under the field`);
    await page.screenshot({ path: path.join(out, `phone360-light-ask-pending-${reduced ? 'reduced' : 'motion'}.png`) });
    release();
    await page.waitForSelector('#answer .used-card', { timeout: 5000 });
    await page.unroute('**/ask?*');
    check(await page.isHidden('#ask-pending') && await page.isEnabled('.ask-form button[type="submit"]'), `ask ux (${tag}): the progress stays after the answer`);
    // The answer is directly under the field (the suggestions come after it), in view, with ONE attribution line.
    const answerBox = await page.locator('#answer .conversation').boundingBox();
    const fieldNow = await page.locator('.ask-field').boundingBox();
    const chipsBox = await page.locator('.chips-row').boundingBox();
    const viewport = page.viewportSize();
    check(answerBox.y > fieldNow.y && answerBox.y < chipsBox.y && answerBox.y >= 0 && answerBox.y < viewport.height, `ask ux (${tag}): the answer is not directly under the field and in view (${answerBox.y} vs chips ${chipsBox.y})`);
    const text = await page.textContent('#answer');
    check((text.match(/Answered by/g) || []).length === 1 && !text.includes('it read only the cards below') && (await page.locator('#answer .answer-head').count()) === 0,
      `ask ux (${tag}): the answer carries more than one attribution line`);
    const note = await page.textContent('#answer .ask-note');
    check(note.includes('used 1 card') && note.includes('AI can be wrong; check the cards. Not investment advice.') && note.includes('How models did in our evaluation'), `ask ux (${tag}): the note says ${note}`);
    check((await sideways(page)) <= 0, `ask ux (${tag}): the answer scrolls sideways at 360 px`);
    await page.screenshot({ path: path.join(out, `phone360-light-ask-answered-${reduced ? 'reduced' : 'motion'}.png`), fullPage: true });
    // A chip is answered with "Looking it up…" and a command's answer says it needs no model.
    await page.route('**/ask?*', async (route) => { await new Promise((r) => setTimeout(r, 400)); await route.continue(); });
    await page.click('.chips-row .chip >> nth=0');
    await page.waitForSelector('#ask-pending.htmx-request', { state: 'visible', timeout: 5000 });
    check((await page.textContent('#ask-pending-label')) === 'Looking it up…', `ask ux (${tag}): a chip's pending answer says ${await page.textContent('#ask-pending-label')}`);
    await page.waitForSelector('#answer .answer', { timeout: 5000 });
    await page.click('details.ask-howto summary');
    check(await page.isVisible('#ask-help'), `ask ux (${tag}): "How this works" does not open`);
    await ctx.close();
  }
}


// withoutTurnstile checks that the first UI, which has no Turnstile, has no widget and loads nothing from Cloudflare.
async function withoutTurnstile(browser) {
  const ctx = await context(browser, { viewport: { width: 360, height: 780 }, colorScheme: 'light', isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  const outside = [];
  page.on('request', (r) => {
    if (!r.url().startsWith(base)) {
      outside.push(r.url());
    }
  });
  const response = await page.goto(base + '/test/ask-model?tiers=T0,T2', { waitUntil: 'networkidle' });
  check(!(response.headers()['content-security-policy'] || '').includes('cloudflare'), 'no turnstile: the page\'s policy names Cloudflare');
  check((await page.locator('.cf-turnstile').count()) === 0 && outside.length === 0, `no turnstile: the page shows a widget or loads ${outside.join(', ')}`);
  await ctx.close();
}

// present waits, in real time, for a selector to match: Playwright's own wait polls on the page's timers, which a fake clock holds.
async function present(page, selector) {
  for (let i = 0; i < 100; i++) {
    if ((await page.locator(selector).count()) > 0) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  throw new Error(`${selector} did not appear`);
}

// settle lets htmx finish a request under a fake clock: its response is handled on timers the clock holds, so the clock runs a little,
// and the network gets real time, a few times over.
async function settle(ctx) {
  for (let i = 0; i < 4; i++) {
    await new Promise((resolve) => setTimeout(resolve, 250));
    await ctx.clock.runFor(250);
  }
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
// passkeyCheck creates a script token while the fake core wants a recent passkey check (D-183): the Access page asks to
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
  await page.goto(base + '/access', { waitUntil: 'networkidle' });
  await page.fill('#token-name', 'nightly');
  await Promise.all([page.waitForLoadState('networkidle'), page.click('form[action="/access/tokens"] button[type="submit"]')]);
  const confirm = page.locator('[data-webauthn="check"]');
  check(await confirm.isVisible(), 'passkey check: no "Confirm with your passkey" after a refused action');
  await page.screenshot({ path: path.join(out, 'desktop-light-passkey-check.png'), fullPage: true });
  // The page is already at /access/tokens, the refused post's answer, so wait for the post again, not the URL.
  const posted = page.waitForResponse((r) => r.url() === base + '/access/tokens' && r.request().method() === 'POST', { timeout: 30_000 });
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

// people checks the Access page's People section (D-210): the invitation form asks for the passkey check like a new token
// (the fake core's check was passed above), shows the one-time link, and Revoke asks for a confirmation first; the section
// does not scroll sideways at 360 px.
async function people(browser) {
  const ctx = await context(browser, { viewport: { width: 360, height: 780 }, colorScheme: 'light', isMobile: true, hasTouch: true }, true);
  const page = await ctx.newPage();
  await page.goto(base + '/access', { waitUntil: 'networkidle' });
  const section = await page.textContent('#people-title ~ .table');
  check(section.includes('Alice') && section.includes('no access yet') && section.includes('revoked'), 'people: the section lacks Alice, Bob\'s "no access yet" or Carol\'s revoked mark');
  check(await sideways(page) <= 0, 'people: the Access page scrolls sideways at 360 px');
  const pending = await page.textContent('#pending-invites-title ~ .table');
  check(pending.includes('Dana') && pending.includes('Cancel invitation') && !pending.includes('#code='), 'people: the pending invitation lacks Dana or its cancel button, or shows a link');
  const history = await page.textContent('#history-title ~ .table');
  check(history.includes('Access granted') && history.includes('Alice') && history.includes('Invitation cancelled'), 'history: the section lacks the grants, Alice or the cancelled invitation');
  check(await page.locator('#history-title ~ p a[href="/settings/audit"]').count() === 1, 'history: no link to the whole audit log');
  await page.fill('#viewer-name', 'Dana');
  await Promise.all([page.waitForLoadState('networkidle'), page.click('form[action="/access/people"] button[type="submit"]')]);
  let body = await page.textContent('body');
  check(body.includes('Invitation for Dana created') && body.includes('/enroll#code='), 'people: no invitation link after inviting a viewer');
  check(await sideways(page) <= 0, 'people: the invitation link scrolls the page sideways at 360 px');
  await page.screenshot({ path: path.join(out, 'phone360-light-people-invited.png'), fullPage: true });
  await page.goto(base + '/access', { waitUntil: 'networkidle' });
  await Promise.all([page.waitForLoadState('networkidle'), page.click('form[action$="/revoke"] button.danger')]);
  body = await page.textContent('body');
  check(body.includes('Revoke Alice?') && body.includes('Yes, revoke Alice'), 'people: Revoke did not ask for a confirmation');
  await page.screenshot({ path: path.join(out, 'phone360-light-people-revoke.png'), fullPage: true });
  await ctx.close();
}

// auditLog checks Settings' Audit log (D-211) at phone and desktop width: the filter applies from its form, a viewer's
// row is named, and the table does not scroll the page sideways.
async function auditLog(browser) {
  for (const [label, options] of [['phone', { viewport: { width: 360, height: 780 }, isMobile: true, hasTouch: true }], ['desktop', { viewport: { width: 1280, height: 800 } }]]) {
    const ctx = await context(browser, options, true);
    const page = await ctx.newPage();
    await page.goto(base + '/settings/audit', { waitUntil: 'networkidle' });
    const all = await page.locator('tbody tr').count();
    check(all === 12, `audit ${label}: ${all} rows, want the fake log's 12`);
    check((await page.textContent('tbody')).includes('Alice'), `audit ${label}: the viewer's sign-in is not named`);
    await page.selectOption('#audit-outcome', 'refused');
    await Promise.all([page.waitForLoadState('networkidle'), page.click('form[aria-label="Filter the audit log"] button[type="submit"]')]);
    const refused = await page.locator('tbody tr').count();
    check(refused === 2, `audit ${label}: ${refused} refused rows, want 2`);
    check(await sideways(page) <= 0, `audit ${label}: scrolls sideways`);
    await page.screenshot({ path: path.join(out, `${label}-audit-filtered-by-form.png`), fullPage: true });
    // While changes' entries wait to enter the log, the page says so, in the item's words, without scrolling sideways.
    check(await page.locator('[data-audit-behind]').count() === 0, `audit ${label}: a banner without an audit-log-behind item`);
    await page.goto(base + '/test/audit-behind?n=3', { waitUntil: 'networkidle' });
    const banner = await page.locator('[data-audit-behind]').allTextContents();
    check(banner.length === 1 && banner[0].startsWith('3 changes are waiting to enter the audit log; OpenFief keeps retrying.'), `audit ${label}: the banner says ${JSON.stringify(banner)}`);
    check(await sideways(page) <= 0, `audit ${label}: the banner scrolls sideways`);
    await page.screenshot({ path: path.join(out, `${label}-audit-behind.png`), fullPage: true });
    await page.goto(base + '/test/audit-behind?n=0', { waitUntil: 'networkidle' });
    check(await page.locator('[data-audit-behind]').count() === 0, `audit ${label}: the banner stays after the item is gone`);
    await ctx.close();
  }
}

// viewerEmpty checks a viewer the owner granted nothing yet (D-210): every page they have says plainly that nothing has
// been shared with them, with no chart and no error, at phone and desktop width.
async function viewerEmpty(browser) {
  for (const [label, options] of [['phone', { viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true }], ['desktop', { viewport: { width: 1280, height: 800 } }]]) {
    const ctx = await browser.newContext(options);
    const page = await ctx.newPage();
    await page.goto(base + '/test/sign-in-viewer');
    for (const target of ['/', '/accounts', '/sleeves', '/inbox']) {
      const response = await page.goto(base + target, { waitUntil: 'networkidle' });
      check(response && response.status() === 200, `viewer ${label} ${target}: status ${response && response.status()}`);
      check((await page.textContent('body')).includes('Nothing has been shared with you yet'), `viewer ${label} ${target}: no empty state`);
      check(await page.locator('canvas').count() === 0, `viewer ${label} ${target}: an empty chart`);
      check(await sideways(page) <= 0, `viewer ${label} ${target}: scrolls sideways`);
      await page.screenshot({ path: path.join(out, `${label}-viewer-empty${target === '/' ? '-overview' : target.replace('/', '-')}.png`), fullPage: true });
    }
    await ctx.close();
  }
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
const turnstileServer = await serve(turnstileAddr, { OPENFIEF_UI_TURNSTILE: '1' });
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
  await askModelBox(browser);
  await withoutTurnstile(browser);
  await turnstileBox(browser);
  await turnstileOnDemand(browser);
  await askProgress(browser);
  await recordPages(browser);
  await passkeyCheck(browser);
  await people(browser);
  await auditLog(browser);
  await viewerEmpty(browser);
  await app();
  await keptOverview();
} finally {
  await browser.close();
  process.kill(-server.pid);
  process.kill(-turnstileServer.pid);
}
if (failures.length > 0) {
  console.error(`${failures.length} browser check(s) failed`);
  process.exit(1);
}
console.log(`browser checks passed; screenshots in ${out}`);
