#!/usr/bin/env bash
# Writes kibi's coverage of every requirement as one JSON document, {"rows": [...]}, the rows tools/docgen reads
# (docs/dev/docs.md#build-and-publishing, docs/dev/kibi.md#setup-notes).
#
# Paging: kibi stops any single query at 120 seconds, and a coverage query costs a fixed ~30 s (kibi's quadratic
# contradiction-witness step) plus time for every row, so the whole report in one query sits close to the limit.
# This asks for KIBI_COVERAGE_PAGE requirements at a time (default 300) and joins the pages' rows. Each page runs
# through kibi-timed.sh, which prints its duration and warns from 90 s, so the guard applies to every page.
# A cheap query first attaches the store, so attaching it, and the local patch's dedupe of a store kibi sync just
# published with duplicates, isn't counted in the first page's limit.
#
# Cache: the report is fully determined by the knowledge base and the kibi that reads it, so outside CI it is kept
# under site/.gen/cache/kibi-coverage, named by a key that hashes the working-tree contents of every .kb file Git
# tracks or would track (kibi reads the working tree, not the index), the installed kibi-cli and kibi-core
# versions, both local kibi patches and this script. A run whose key matches a kept report copies it without
# calling kibi. Under CI (CI or GITHUB_ACTIONS is true), with KIBI_COVERAGE_CACHE=off, or with ingested proof runs
# in .kb/proof/runs (they count only for the workspace snapshot they were made from), the cache is neither read nor
# written: the published site always comes from a fresh report. A report kibi answered from a store that wasn't
# fresh (meta.syncState) is written but not kept.
#
# Usage: kibi-coverage.sh <kibi> <out.json>   (run from the repository root)
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: kibi-coverage.sh <kibi> <out.json>" >&2
  exit 2
fi
kibi="$1"
out="$2"
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
page="${KIBI_COVERAGE_PAGE:-300}"
cache_dir="site/.gen/cache/kibi-coverage"
keep=3
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# cache_off names why the cache is skipped, or prints nothing when it is used.
cache_off() {
  if [ "${CI:-}" = true ] || [ "${GITHUB_ACTIONS:-}" = true ]; then
    echo "CI always regenerates"
  elif [ "${KIBI_COVERAGE_CACHE:-}" = off ]; then
    echo "KIBI_COVERAGE_CACHE=off"
  elif [ -n "$(find .kb/proof/runs -type f -print -quit 2>/dev/null)" ]; then
    # Ingested proof runs count only for the workspace snapshot they were made from, which this key doesn't cover.
    echo "ingested proof runs in .kb/proof/runs"
  fi
}

# version prints an installed npm package's version, or "missing".
version() {
  local manifest="node_modules/$1/package.json"
  if [ -f "$manifest" ]; then
    node -p 'require(process.argv[1]).version' "$PWD/$manifest"
  else
    echo missing
  fi
}

# cache_key hashes everything the report depends on: one line per input, then sha256 of the lot.
cache_key() {
  {
    echo "kibi-cli $(version kibi-cli)"
    echo "kibi-core $(version kibi-core)"
    # Tracked and untracked-but-not-ignored .kb files: kibi reads both.
    {
      git ls-files -z --cached --others --exclude-standard -- .kb
      printf '%s\0' patches/kibi-cli+*.patch patches/kibi-core+*.patch
    } | sort -zu >"$work/inputs"
    : >"$work/present"
    while IFS= read -r -d '' f; do
      if [ -f "$f" ]; then printf '%s\0' "$f" >>"$work/present"; else echo "missing $f"; fi
    done <"$work/inputs"
    xargs -0 -r sha256sum <"$work/present"
    sha256sum <"$here/kibi-coverage.sh" | sed 's|-$|scripts/kibi-coverage.sh|'
  } | sha256sum | cut -d' ' -f1
}

off="$(cache_off)"
key=""
if [ -z "$off" ]; then
  key="$(cache_key)"
  if [ -f "$cache_dir/$key.json" ]; then
    cp "$cache_dir/$key.json" "$out"
    touch "$cache_dir/$key.json"
    echo "kibi coverage: reused cached report for $key" >&2
    exit 0
  fi
  echo "kibi coverage: no cached report for $key, asking kibi" >&2
else
  echo "kibi coverage: cache skipped ($off), asking kibi" >&2
fi

"$kibi" query req --limit 1 >/dev/null

offset=0
n=0
while :; do
  bash "$here/kibi-timed.sh" "$kibi" coverage --by req --include-passing --limit "$page" --offset "$offset" \
    --format json >"$work/page-$n.json"
  # The page's row count and kibi's count of all requirements in the report (summary.total).
  node -p 'const d = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"));
    `${d.rows.length} ${d.summary.total}`' "$work/page-$n.json" >"$work/page-$n.txt"
  read -r rows total <"$work/page-$n.txt"
  echo "kibi coverage: page $((n + 1)) (offset $offset) has $rows rows" >&2
  n=$((n + 1))
  if [ "$rows" -lt "$page" ] || [ $((offset + rows)) -ge "$total" ]; then
    break
  fi
  offset=$((offset + page))
done

# Join the pages' rows one page at a time, so only one page is parsed at once; print the row count and whether every
# page came from a fresh store. Pages that don't add up to kibi's total, each requirement once, fail the build.
node - "$work/joined.json" "$n" "$work" "$total" >"$work/joined.txt" <<'EOF'
const fs = require("fs");
const [outFile, pages, dir, total] = process.argv.slice(2);
const fd = fs.openSync(outFile, "w");
fs.writeSync(fd, '{"rows":[');
let first = true;
let fresh = true;
const ids = new Set();
let count = 0;
for (let i = 0; i < Number(pages); i++) {
  const doc = JSON.parse(fs.readFileSync(`${dir}/page-${i}.json`, "utf8"));
  if (!doc.meta || doc.meta.syncState !== "fresh") fresh = false;
  for (const row of doc.rows) {
    fs.writeSync(fd, (first ? "" : ",") + JSON.stringify(row));
    first = false;
    ids.add(row.id);
    count++;
  }
}
fs.writeSync(fd, "]}\n");
fs.closeSync(fd);
if (count !== Number(total) || ids.size !== count) {
  console.error(`kibi coverage: the pages hold ${count} rows for ${ids.size} requirements, but kibi counts ${total}`);
  process.exit(1);
}
console.log(count, fresh ? "yes" : "no");
EOF
read -r count fresh <"$work/joined.txt"
mv "$work/joined.json" "$out"
echo "kibi coverage: $count requirements in $n pages of up to $page" >&2

if [ -n "$key" ]; then
  if [ "$fresh" = yes ]; then
    mkdir -p "$cache_dir"
    cp "$out" "$cache_dir/.$key.json.tmp"
    mv "$cache_dir/.$key.json.tmp" "$cache_dir/$key.json"
    # Keep only the newest few reports.
    ls -t "$cache_dir"/*.json | tail -n +$((keep + 1)) | xargs -r rm -f
    echo "kibi coverage: regenerated and cached the report for $key" >&2
  else
    echo "kibi coverage: regenerated the report; not cached, kibi's store wasn't fresh (kibi sync)" >&2
  fi
fi
