#!/usr/bin/env bash
# The pre-commit hook's runner (task check:commit; docs/dev/pipeline.md#affected-checks). It writes the staged
# changes and the Go package listings to a temporary directory, asks tools/affected which task targets they need
# (the tool prints what it selected and why), and runs them in order, stopping at the first failure. FULL, and a
# merge being committed, select the full check, task check:commit:full. If the selector itself fails, so does
# the commit. With DRY_RUN=1 it only prints the plan (task check:plan).
# Usage: check-commit.sh <task executable> "<Go module dirs>" [FULL]
set -euo pipefail

task_exe="$1"
modules="$2"
full="${3:-}"
here="$(cd "$(dirname "$0")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

reason=""
if [ -n "$full" ]; then
  reason="FULL is set"
elif git rev-parse -q --verify MERGE_HEAD >/dev/null; then
  reason="a merge is being committed: what it brings in is checked as a whole"
fi

if [ -n "$reason" ]; then
  go run ./tools/affected hook -full "$reason" >"$work/plan"
else
  git diff --cached --name-status --no-renames -z >"$work/changes"
  read -r -a module_dirs <<<"$modules"
  bash "$here/go-packages.sh" "$work" "${module_dirs[@]}"
  go run ./tools/affected hook -in "$work" -modules "$modules" >"$work/plan"
fi

if [ "${DRY_RUN:-}" = 1 ]; then
  exit 0
fi

# Each plan line is a task and its variables, tab-separated. The plan is read on its own descriptor, so a task
# that reads standard input can't swallow the lines after it.
while IFS=$'\t' read -r -a step <&3; do
  [ "${#step[@]}" -gt 0 ] || continue
  "$task_exe" "${step[@]}"
done 3<"$work/plan"
