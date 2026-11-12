#!/usr/bin/env bash
# The pre-commit hook's runner (task check:commit; docs/dev/pipeline.md#affected-checks). It writes the staged
# changes and the Go package listings to a temporary directory, asks tools/affected which task targets they need
# (the tool prints what it selected and why), and runs them in order, stopping at the first failure, and says how
# long each took and the total. A merge being committed needs no special case: what is staged then is the merge's
# diff against its first parent (HEAD), so the selection covers what the merge brings in. Only FULL selects the
# full check, task check:commit:full. If the selector itself fails, so does the commit. With DRY_RUN=1 it only
# prints the plan (task check:plan).
# Usage: check-commit.sh <task executable> "<Go module dirs>" [FULL]
set -euo pipefail
# SECONDS counts from here, so the closing total includes planning.
SECONDS=0

task_exe="$1"
modules="$2"
full="${3:-}"
here="$(cd "$(dirname "$0")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"; [ "${DRY_RUN:-}" = 1 ] || printf "check-commit: total %ds\n" "$SECONDS" >&2' EXIT

if [ -n "$full" ]; then
  go run ./tools/affected hook -full "FULL is set" >"$work/plan"
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
  started=$SECONDS
  status=0
  "$task_exe" "${step[@]}" || status=$?
  printf 'check-commit: %s: %ds\n' "${step[*]}" "$((SECONDS - started))" >&2
  [ "$status" -eq 0 ] || exit "$status"
done 3<"$work/plan"
