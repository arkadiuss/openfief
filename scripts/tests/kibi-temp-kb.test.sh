#!/usr/bin/env bash
# Test of the kibi-cli patch that keeps the staged check's throwaway KB off the disk (patches/kibi-cli+2.42.0.patch;
# docs/dev/kibi.md): projecting entities into it must save nothing, because each save rewrites the whole store and
# made `kibi check --staged` quadratic in the KB's size (kibi: TEST-tooling-kibi-temp-kb). Runs kibi's own createTempKb
# and projectStagedEntities on a few hundred synthetic entities, so the projection spans several batches.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
dist="$root/node_modules/kibi-cli/dist"
[ -d "$dist" ] || { echo "kibi-temp-kb.test: kibi-cli is not installed (task kibi:install)" >&2; exit 1; }

node --input-type=module - "$dist" <<'JS'
import { existsSync, statSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const dist = process.argv[2];
const { createTempKb, cleanupTempKb, projectStagedEntities } = await import(pathToFileURL(path.join(dist, "traceability/temp-kb.js")).href);

const count = 300;
const results = Array.from({ length: count }, (_, i) => ({
  entity: {
    id: `REQ-synthetic-${i}`,
    type: "req",
    title: `Synthetic requirement ${i}`,
    status: "active",
    created_at: "2026-01-01T00:00:00.000Z",
    updated_at: "2026-01-01T00:00:00.000Z",
    source: `.kb/requirements/REQ-synthetic-${i}.md`,
  },
  relationships: [],
  sourceFile: `.kb/requirements/REQ-synthetic-${i}.md`,
}));

const ctx = await createTempKb();
let failure = null;
try {
  await projectStagedEntities(ctx.prolog, results);
  const rdf = path.join(ctx.kbPath, "kb.rdf");
  const saved = existsSync(rdf) ? statSync(rdf).size : 0;
  if (saved > 0) failure = `projecting ${count} entities wrote ${saved} bytes to the throwaway KB's kb.rdf; it must save nothing`;
  const found = await ctx.prolog.query("aggregate_all(count, kb_entity(_, req, _), N)");
  if (!found.success || Number(found.bindings.N) !== count) failure ??= `the projected KB holds ${found.bindings?.N} requirements, want ${count}`;
} finally {
  await cleanupTempKb(ctx.tempDir);
}
if (existsSync(ctx.tempDir)) failure ??= "cleanupTempKb left its directory behind";
if (failure) {
  console.error(`FAIL: ${failure}`);
  process.exit(1);
}
JS
echo "kibi-temp-kb.test: ok"
