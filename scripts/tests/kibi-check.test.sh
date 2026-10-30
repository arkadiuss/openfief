#!/usr/bin/env bash
# Tests for scripts/kibi-check.sh with a fake kibi whose check answers come from a list, one per call; no real
# knowledge base is touched (kibi: TEST-tooling-kibi-check).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# The fake kibi logs each call; `check` takes the next line of $FAKE_ANSWERS: ok, missing (a compile defect),
# duplicate (the paginated projection's) or authored (any other violation).
cat >"$work/kibi" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$FAKE_LOG"
[ "$1" = check ] || exit 0
n="$(grep -c '^check' "$FAKE_LOG")"
case "$(sed -n "${n}p" "$FAKE_ANSWERS")" in
  ok) echo "No violations found. KB is valid." ;;
  missing) echo "[source-relationship-parity] Authored relationship is missing from the compiled KB: verified_by SCEN-x->TEST-x"; exit 1 ;;
  duplicate) echo "Full KB paginated entity projection returned duplicate entity ID: SYM-x"; exit 1 ;;
  *) echo "[symbol-coverage] Production symbol lacks qualifying requirement coverage"; exit 1 ;;
esac
EOF
chmod +x "$work/kibi"

# gate runs kibi-check.sh with the fake's check answers given as arguments.
gate() {
  : >"$work/log"
  printf '%s\n' "$@" >"$work/answers"
  FAKE_LOG="$work/log" FAKE_ANSWERS="$work/answers" bash "$root/scripts/kibi-check.sh" "$work/kibi" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" "$work/log" >&2
    failures=$((failures + 1))
  fi
}

# calls prints the fake's calls on one line.
calls() { paste -sd, "$work/log"; }

gate ok && rc=0 || rc=$?
check "a valid KB passes with one check" "$([ "$rc" -eq 0 ] && [ "$(calls)" = check ] && echo yes || echo no)"

gate authored && rc=0 || rc=$?
check "an authored violation fails at once, without a recompile" "$([ "$rc" -ne 0 ] && [ "$(calls)" = check ] && echo yes || echo no)"

gate missing ok && rc=0 || rc=$?
check "a compile defect recompiles once and the second check decides" \
  "$([ "$rc" -eq 0 ] && [ "$(calls)" = "check,sync --rebuild,check" ] && echo yes || echo no)"

gate duplicate ok && rc=0 || rc=$?
check "a duplicate projection row counts as a compile defect" "$([ "$rc" -eq 0 ] && [ "$(calls)" = "check,sync --rebuild,check" ] && echo yes || echo no)"

gate missing missing && rc=0 || rc=$?
check "a defect that survives the recompile fails, with no third try" \
  "$([ "$rc" -ne 0 ] && [ "$(calls)" = "check,sync --rebuild,check" ] && echo yes || echo no)"

gate missing authored && rc=0 || rc=$?
check "the recompiled store's own violations fail" "$([ "$rc" -ne 0 ] && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-check.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-check.test.sh: all $cases checks passed"
