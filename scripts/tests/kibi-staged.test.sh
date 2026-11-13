#!/usr/bin/env bash
# Tests for scripts/kibi-staged.sh with a fake kibi that only logs its calls: it removes the throwaway stores of
# dead processes and nothing else, then runs both checks in order, and a failing first check stops the second
# (kibi: TEST-tooling-kibi-staged).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0

cat >"$work/kibi" <<'EOF2'
#!/usr/bin/env bash
echo "$*" >>"$FAKE_LOG"
[ "$1 $FAKE_FAIL" = "check-generated generated" ] && exit 1
exit 0
EOF2
chmod +x "$work/kibi"

expect() {
  if [ "$2" != yes ]; then
    echo "FAIL: $1" >&2
    failures=$((failures + 1))
  fi
}

mkdir "$work/tmp"
sleep 300 &
live=$!
trap 'kill "$live" 2>/dev/null || true; rm -rf "$work"' EXIT
dead=$((live + 100000))
while kill -0 "$dead" 2>/dev/null; do dead=$((dead + 1)); done
mkdir -p "$work/tmp/kibi-precommit-$dead-1-aa/kb" "$work/tmp/kibi-precommit-$live-1-bb/kb" "$work/tmp/other-dir" "$work/tmp/kibi-precommit-x-1-cc"
: >"$work/tmp/kibi-precommit-$dead-1-aa/kb/kb.rdf"

run() {
  : >"$work/log"
  TMPDIR="$work/tmp" FAKE_LOG="$work/log" FAKE_FAIL="${1:-}" bash "$root/scripts/kibi-staged.sh" "$work/kibi"
}

run && rc=0 || rc=$?
expect "a passing run succeeds" "$([ "$rc" -eq 0 ] && echo yes || echo no)"
expect "both checks run, in order" "$([ "$(paste -sd, "$work/log")" = "check-generated --staged --changed-only,check --staged" ] && echo yes || echo no)"
expect "a dead process's store is removed" "$([ ! -e "$work/tmp/kibi-precommit-$dead-1-aa" ] && echo yes || echo no)"
expect "a live process's store stays" "$([ -d "$work/tmp/kibi-precommit-$live-1-bb" ] && echo yes || echo no)"
expect "other directories stay" "$([ -d "$work/tmp/other-dir" ] && [ -d "$work/tmp/kibi-precommit-x-1-cc" ] && echo yes || echo no)"

run generated && rc=0 || rc=$?
expect "a failing generated-code check fails the step" "$([ "$rc" -ne 0 ] && echo yes || echo no)"
expect "and the staged check does not run" "$([ "$(paste -sd, "$work/log")" = "check-generated --staged --changed-only" ] && echo yes || echo no)"

[ "$failures" -eq 0 ] || exit 1
echo "kibi-staged.test: ok"
