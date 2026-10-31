#!/usr/bin/env node
// Renders the built-in UI's app icons from the Seal (design/marks/seal.svg), the design system's mark for an app
// icon, onto the argent ground from design/tokens.css: 192 and 512 px for any use, a 512 px maskable one whose seal
// stays inside the central 80% circle every platform's mask keeps, and Apple's 180 px touch icon.
// Usage: node scripts/app-icons.mjs   (task design:icons); writes design/marks/app-icon-*.png.
// The browser is DOCS_CHROMIUM, else a Playwright-managed Chromium, else the installed Google Chrome.
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright-core';

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const marks = path.join(root, 'design', 'marks');

// The seal's share of the icon's height: the whole tile for any use, the safe zone for a maskable icon.
const icons = [
  { file: 'app-icon-180.png', size: 180, seal: 0.78 },
  { file: 'app-icon-192.png', size: 192, seal: 0.78 },
  { file: 'app-icon-512.png', size: 512, seal: 0.78 },
  { file: 'app-icon-maskable-512.png', size: 512, seal: 0.62 },
];

// token reads a tincture from the first :root block of tokens.css, the one place a color is written.
function token(name) {
  const css = readFileSync(path.join(root, 'design', 'tokens.css'), 'utf8');
  const block = css.slice(css.indexOf(':root {'), css.indexOf('}'));
  const match = block.match(new RegExp(`${name}\\s*:\\s*([^;]+);`));
  if (!match) {
    throw new Error(`tokens.css defines no ${name}`);
  }
  return match[1].trim();
}

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

const seal = readFileSync(path.join(marks, 'seal.svg'), 'utf8');
const ground = token('--argent');
const browser = await chromium.launch(browserOptions());
try {
  for (const icon of icons) {
    const page = await browser.newPage({ viewport: { width: icon.size, height: icon.size }, deviceScaleFactor: 1 });
    const height = Math.round(icon.size * icon.seal);
    await page.setContent(`<!doctype html><html><body style="margin:0;width:${icon.size}px;height:${icon.size}px;` +
      `display:grid;place-items:center;background:${ground}">` +
      `<img style="height:${height}px;width:auto;display:block" src="data:image/svg+xml;base64,${Buffer.from(seal).toString('base64')}">` +
      '</body></html>');
    await page.waitForFunction(() => document.images[0].complete);
    await page.screenshot({ path: path.join(marks, icon.file), omitBackground: false });
    await page.close();
    console.log(`wrote design/marks/${icon.file}`);
  }
} finally {
  await browser.close();
}
