#!/usr/bin/env bash
# Writes kibi's coverage of every requirement as one JSON document, {"rows": [...]}, the rows tools/docgen reads.
# kibi answers one query in at most 120 seconds and the whole knowledge base no longer fits in one, so this asks
# for it a page at a time and joins the pages' rows.
# Usage: kibi-coverage.sh <kibi> <out.json>   (KIBI_COVERAGE_PAGE sets the page size, default 500)
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: kibi-coverage.sh <kibi> <out.json>" >&2
  exit 2
fi
kibi="$1"
out="$2"
page="${KIBI_COVERAGE_PAGE:-500}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

offset=0
n=0
while :; do
  "$kibi" coverage --by req --include-passing --limit "$page" --offset "$offset" --format json >"$work/page-$n.json"
  rows="$(gojq '.rows | length' "$work/page-$n.json")"
  n=$((n + 1))
  if [ "$rows" -lt "$page" ]; then
    break
  fi
  offset=$((offset + page))
done

gojq -s '{rows: (map(.rows) | add)}' "$work"/page-*.json >"$out"
echo "kibi-coverage: $(gojq '.rows | length' "$out") requirements in $n pages"
