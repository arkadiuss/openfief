#!/usr/bin/env bash
# The pre-commit hook's kibi step (task kibi:staged; docs/dev/pipeline.md#the-kibi-gate). First it removes the
# throwaway knowledge bases that earlier `kibi check --staged` runs left in the temporary directory: kibi builds
# one per run (kibi-precommit-<pid>-<time>-<id>, tens of megabytes) and removes it on exit, but a run that was
# killed (a timeout, Ctrl-C, a restarted container) leaves it behind. Only directories whose process is gone are
# removed, so a check running in another checkout is never touched. Then it runs the two checks, the second only
# if the first passes. Nothing about them is skipped or loosened.
# Usage: kibi-staged.sh <kibi executable>
set -uo pipefail

kibi="$1"
tmp="${TMPDIR:-/tmp}"

for dir in "$tmp"/kibi-precommit-*; do
  [ -d "$dir" ] && [ -O "$dir" ] || continue
  name="${dir##*/}"
  pid="${name#kibi-precommit-}"
  pid="${pid%%-*}"
  case "$pid" in '' | *[!0-9]*) continue ;; esac
  kill -0 "$pid" 2>/dev/null && continue
  rm -rf "$dir"
done

"$kibi" check-generated --staged --changed-only || exit $?
"$kibi" check --staged
