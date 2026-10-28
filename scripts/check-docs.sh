#!/usr/bin/env bash
# ID checks for the design documents (see .claude/skills/design-docs).
# DOCS_BASE sets the revision ground rules are compared against (default HEAD; CI uses the merge base).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root/docs"
base="${DOCS_BASE:-HEAD}"
files=(DESIGN.md glossary.md research.md proposals/*.md)
fail=0

defined() { grep -ohE "^ *- \*\*\[$1-[0-9]+\]" "${files[@]}" | grep -oE "$1-[0-9]+" | sort; }
referenced() { grep -ohE "\b$1-[0-9]+\b" "${files[@]}" | sort -u; }

for tag in GR D P T OQ R; do
  missing="$(comm -23 <(referenced "$tag") <(defined "$tag" | uniq))"
  if [ -n "$missing" ]; then
    echo "docs: undefined $tag tags: $(echo "$missing" | tr '\n' ' ')" >&2
    fail=1
  fi
  twice="$(defined "$tag" | uniq -d)"
  if [ -n "$twice" ]; then
    echo "docs: $tag IDs defined more than once: $(echo "$twice" | tr '\n' ' ')" >&2
    fail=1
  fi
done

if git diff "$base" -U0 -- DESIGN.md | grep -qE '^[-+]- \*\*\[GR-'; then
  echo "docs: ground-rule lines in docs/DESIGN.md changed against $base; ground rules are never edited" >&2
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "docs: ID checks passed"
fi
exit "$fail"
