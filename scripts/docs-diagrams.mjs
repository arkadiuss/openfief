#!/usr/bin/env node
// Renders every Mermaid block the docs site collected (its first build pass writes diagrams.json) to a light
// and a dark SVG, so a broken diagram fails the build and readers' browsers run no JavaScript (docs/dev/docs.md).
// Usage: node scripts/docs-diagrams.mjs <diagrams.json> <out-dir>
// The browser is DOCS_CHROMIUM, else a Playwright-managed Chromium, else the installed Google Chrome.
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright-core';

const require = createRequire(import.meta.url);
const mermaidScript = require.resolve('mermaid/dist/mermaid.min.js');
const mermaidVersion = require('mermaid/package.json').version;
// The design system's palette (docs/dev/design-system.md, design/tokens.css): vert nodes, gold notes, sable text.
// On the dark ground, node borders and icons use the light vert tint, which reads where --vert-bright is too dim.
// Pie slices are light tints so one dark label color reads on every slice in both schemes.
const pie = { pie1: '#8ccba6', pie2: '#d9ae4a', pie3: '#cbd2c9', pie4: '#c99a6e', pieSectionTextColor: '#14171a', pieStrokeColor: '#14171a', pieOuterStrokeColor: '#9aa59e' };
const themes = {
  light: {
    theme: 'base',
    themeVariables: {
      ...pie, background: '#ffffff', fontSize: '15px',
      primaryColor: '#e3ece5', primaryBorderColor: '#1e3a2b', primaryTextColor: '#14171a',
      secondaryColor: '#f4e6c4', secondaryBorderColor: '#c99a2e', tertiaryColor: '#f3f4ef', tertiaryBorderColor: '#9aa59e',
      lineColor: '#3e4744', textColor: '#14171a', clusterBkg: '#f3f4ef', clusterBorder: '#9aa59e', edgeLabelBackground: '#ffffff',
      noteBkgColor: '#f4e6c4', noteBorderColor: '#c99a2e', pieTitleTextColor: '#14171a', pieLegendTextColor: '#14171a',
    },
  },
  dark: {
    theme: 'base',
    themeVariables: {
      ...pie, darkMode: true, background: '#18201c', fontSize: '15px',
      primaryColor: '#22302a', primaryBorderColor: '#8ccba6', primaryTextColor: '#e8ece8',
      secondaryColor: '#3a3016', secondaryBorderColor: '#d9ae4a', tertiaryColor: '#18201c', tertiaryBorderColor: '#4f5f56',
      lineColor: '#a9b5ad', textColor: '#e8ece8', clusterBkg: '#18201c', clusterBorder: '#4f5f56', edgeLabelBackground: '#18201c',
      noteBkgColor: '#3a3016', noteBorderColor: '#d9ae4a', noteTextColor: '#e8ece8', pieTitleTextColor: '#e8ece8', pieLegendTextColor: '#e8ece8',
    },
  },
};
// iconPack turns the design system's icons (design/icons/*.svg) into a Mermaid icon pack named fief, so a diagram
// draws one as fief:<file name>, such as A@{ icon: "fief:connector", label: "IBKR Flex" }. Each file's drawing is
// used as it is: strokes in currentColor, which Mermaid sets to the node's border color, and the gold fief accent.
function iconPack(dir) {
  const icons = {};
  for (const file of readdirSync(dir).filter((f) => f.endsWith('.svg')).sort()) {
    const svg = readFileSync(path.join(dir, file), 'utf8');
    const body = svg.replace(/^[\s\S]*?<svg[^>]*>/, '').replace(/<\/svg>\s*$/, '').replace(/<title>[\s\S]*?<\/title>/, '');
    icons[file.slice(0, -'.svg'.length)] = { body };
  }
  return { prefix: 'fief', width: 24, height: 24, icons };
}
const icons = iconPack(path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'design', 'icons'));
// The icons' gold is --or; on a dark ground the design system uses --or-bright.
const darkGold = (svg) => svg.replaceAll('#C99A2E', '#D9AE4A');
// The cache of rendered SVGs is valid for one Mermaid version, one set of themes and one set of icons.
const cacheKey = `${mermaidVersion} ${JSON.stringify(themes)} ${JSON.stringify(icons)}`;
const fontFamily = 'system-ui, -apple-system, "Segoe UI", Roboto, sans-serif';

// browserOptions picks the Chromium to drive, without downloading one.
function browserOptions() {
  if (process.env.DOCS_CHROMIUM) {
    return { executablePath: process.env.DOCS_CHROMIUM };
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

// uniqueDiagrams groups the collected blocks by hash, keeping every page a diagram appears on for error messages.
function uniqueDiagrams(collected) {
  const byHash = new Map();
  for (const { hash, code, page } of collected) {
    const entry = byHash.get(hash) ?? { hash, code, pages: [] };
    if (!entry.pages.includes(page)) {
      entry.pages.push(page);
    }
    byHash.set(hash, entry);
  }
  return [...byHash.values()];
}

// renderAll renders each missing SVG and returns the failures; SVGs already in outDir for this cache key are kept.
async function renderAll(diagrams, outDir) {
  const versionFile = path.join(outDir, '.mermaid-version');
  if (!existsSync(versionFile) || (await readFile(versionFile, 'utf8')) !== cacheKey) {
    await rm(outDir, { recursive: true, force: true });
  }
  await mkdir(outDir, { recursive: true });
  await writeFile(versionFile, cacheKey);

  const todo = diagrams.flatMap((d) => Object.keys(themes).map((variant) => ({ ...d, variant })))
    .filter((job) => !existsSync(path.join(outDir, `${job.hash}-${job.variant}.svg`)));
  if (todo.length === 0) {
    return [];
  }
  const browser = await chromium.launch(browserOptions());
  const failures = [];
  try {
    const page = await browser.newPage();
    await page.setContent('<!doctype html><html><body></body></html>');
    await page.addScriptTag({ path: mermaidScript });
    await page.evaluate((pack) => globalThis.mermaid.registerIconPacks([{ name: 'fief', icons: pack }]), icons);
    for (const job of todo) {
      const result = await page.evaluate(async ({ code, id, theme, font }) => {
        try {
          globalThis.mermaid.initialize({ startOnLoad: false, securityLevel: 'strict', ...theme, fontFamily: font });
          const { svg } = await globalThis.mermaid.render(id, code);
          return { svg };
        } catch (error) {
          return { error: String(error?.message ?? error) };
        }
      }, { code: job.code, id: `m${job.hash}${job.variant[0]}`, theme: themes[job.variant], font: fontFamily });
      if (result.error || !result.svg || result.svg.includes('Syntax error')) {
        failures.push({ ...job, error: result.error ?? 'Mermaid returned an error diagram' });
        continue;
      }
      await writeFile(path.join(outDir, `${job.hash}-${job.variant}.svg`), job.variant === 'dark' ? darkGold(result.svg) : result.svg);
    }
  } finally {
    await browser.close();
  }
  return failures;
}

const [collectedFile, outDir] = process.argv.slice(2);
if (!collectedFile || !outDir) {
  console.error('usage: docs-diagrams.mjs <diagrams.json> <out-dir>');
  process.exit(2);
}
const diagrams = uniqueDiagrams(JSON.parse(await readFile(collectedFile, 'utf8')));
const failures = await renderAll(diagrams, outDir);
for (const f of failures.filter((x) => x.variant === 'light')) {
  console.error(`docs-diagrams: ${f.pages.join(', ')}: diagram starting "${f.code.split('\n')[0]}" does not render: ${f.error}`);
}
if (failures.length > 0) {
  process.exit(1);
}
console.log(`docs-diagrams: ${diagrams.length} diagrams rendered (Mermaid ${mermaidVersion})`);
