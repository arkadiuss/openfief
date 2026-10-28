#!/usr/bin/env bash
# Fails when generated files differ from what was staged or committed, after `task gen` regenerated them.
# Checks the working tree against the index, so it works both in the pre-commit hook and on a clean CI checkout.
# Usage: gen-check.sh <generated path>...
set -euo pipefail

if [ "$#" -eq 0 ]; then
  echo "gen-check: no paths given" >&2
  exit 2
fi

stale="$(git diff --name-only -- "$@")"
untracked="$(git ls-files --others --exclude-standard -- "$@")"
if [ -n "$stale$untracked" ]; then
  echo "gen-check: generated files are stale or not added; run \`task gen\` and stage them:" >&2
  printf '%s\n' $stale $untracked >&2
  exit 1
fi
echo "gen-check: generated files are up to date"
