#!/usr/bin/env bash
# Provenance checks for docs/DESIGN.md and docs/SPEC.md (see .claude/skills/design-docs).
# DOCS_BASE sets the revision ground rules are compared against (default HEAD; CI uses the merge base).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root/docs"
base="${DOCS_BASE:-HEAD}"
fail=0

defined() { grep -oE "^- \*\*\[$1-[0-9]+\]" DESIGN.md | grep -oE "$1-[0-9]+" | sort -u; }
referenced() { grep -ohE "\b$1-[0-9]+\b" DESIGN.md SPEC.md | sort -u; }

for tag in GR D AC R P T OQ; do
  missing="$(comm -23 <(referenced "$tag") <(defined "$tag"))"
  if [ -n "$missing" ]; then
    echo "docs: undefined $tag tags: $(echo "$missing" | tr '\n' ' ')" >&2
    fail=1
  fi
done

pcc_defined="$({ grep -oE '\*\*\[P-CC-[0-9]+\]' DESIGN.md; grep -oE '^\| P-CC-[0-9]+' SPEC.md; } | grep -oE 'P-CC-[0-9]+' | sort -u)"
pcc_missing="$(comm -23 <(grep -ohE 'P-CC-[0-9]+' DESIGN.md SPEC.md | sort -u) <(echo "$pcc_defined"))"
if [ -n "$pcc_missing" ]; then
  echo "docs: undefined P-CC tags: $(echo "$pcc_missing" | tr '\n' ' ')" >&2
  fail=1
fi

if git diff "$base" -U0 -- DESIGN.md | grep -qE '^[-+]- \*\*\[GR-'; then
  echo "docs: ground-rule lines in docs/DESIGN.md changed against $base; ground rules are never edited" >&2
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "docs: provenance checks passed"
fi
exit "$fail"
