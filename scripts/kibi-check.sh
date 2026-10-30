#!/usr/bin/env bash
# The full knowledge-base gate (task kibi:check; docs/dev/pipeline.md#the-kibi-gate). It runs `kibi check`. kibi
# 2.36's compile (`kibi sync`, `kibi sync --rebuild`) sometimes publishes a store that lacks records the authored
# files hold: under load, about one rebuild in ten on this repository, and a check of that store then fails with
# "Authored relationship is missing from the compiled KB" (or, once in CI, a duplicate row in the paginated entity
# projection), the same on every later check until the store is compiled again. A failure that names such a
# compile defect is therefore not a verdict on the authored records: this script compiles the store once more
# from the same files (`kibi sync --rebuild`) and checks again, and that second check's result is final. Any other
# failure is final at once. A real authored problem, such as an endpoint in a file Git doesn't track, survives the
# recompile and still fails. docs/dev/pipeline.md#not-built-yet describes the kibi defect.
# Usage: kibi-check.sh <kibi executable>
set -uo pipefail

kibi="$1"
out="$(mktemp)"
trap 'rm -f "$out"' EXIT

"$kibi" check >"$out" 2>&1
status=$?
cat "$out"
if [ "$status" -eq 0 ]; then
  exit 0
fi
if ! grep -qE 'Authored relationship is missing from the compiled KB|paginated entity projection returned duplicate entity ID' "$out"; then
  exit "$status"
fi
echo "kibi-check: the compiled store disagrees with the authored files (a kibi compile defect); compiling it again from the same files and checking once more" >&2
"$kibi" sync --rebuild || exit 1
"$kibi" check
