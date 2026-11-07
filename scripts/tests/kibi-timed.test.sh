#!/usr/bin/env bash
# Tests for scripts/kibi-timed.sh with a fake kibi and a fake clock: `date +%s` answers from a list, one line per
# call, so no case waits for real time and no knowledge base is touched (kibi: TEST-tooling-kibi-timed).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# The fake kibi writes a known line to each stream and exits with $FAKE_STATUS.
cat >"$work/kibi" <<'EOF'
#!/usr/bin/env bash
echo '{"rows":[]}'
echo "kibi on stderr: $*" >&2
exit "${FAKE_STATUS:-0}"
EOF
# The fake date takes the next line of $FAKE_CLOCK on each call.
mkdir "$work/bin"
cat >"$work/bin/date" <<'EOF'
#!/usr/bin/env bash
n=$(($(wc -l <"$FAKE_CALLS") + 1))
echo x >>"$FAKE_CALLS"
sed -n "${n}p" "$FAKE_CLOCK"
EOF
chmod +x "$work/kibi" "$work/bin/date"

# timed runs kibi-timed.sh for `kibi coverage --format json` on a clock that starts at 1000 and ends <seconds> later.
# GHA is the value of GITHUB_ACTIONS for the run (empty: not under Actions).
timed() {
  local seconds="$1" gha="$2"
  printf '1000\n%s\n' "$((1000 + seconds))" >"$work/clock"
  : >"$work/calls"
  env GITHUB_ACTIONS="$gha" FAKE_CLOCK="$work/clock" FAKE_CALLS="$work/calls" PATH="$work/bin:$PATH" \
    bash "$root/scripts/kibi-timed.sh" "$work/kibi" coverage --format json >"$work/out" 2>"$work/err"
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  out | /' "$work/out" >&2
    sed 's/^/  err | /' "$work/err" >&2
    failures=$((failures + 1))
  fi
}

# has and lacks look for a fixed string in the run's stderr.
has() { grep -qF -- "$1" "$work/err"; }
lacks() { ! grep -qF -- "$1" "$work/err"; }
# clean_stdout is true when stdout holds exactly the fake kibi's line.
clean_stdout() { [ "$(cat "$work/out")" = '{"rows":[]}' ]; }

timed 30 "" && rc=0 || rc=$?
check "a fast query prints its duration and no warning" \
  "$([ "$rc" -eq 0 ] && has "kibi coverage took 30 s" && lacks warning && clean_stdout && echo yes || echo no)"

timed 89 true && rc=0 || rc=$?
check "89 s under Actions is not a warning yet" "$([ "$rc" -eq 0 ] && lacks warning && echo yes || echo no)"

timed 90 true && rc=0 || rc=$?
check "90 s under Actions is a workflow warning, on stderr only" \
  "$([ "$rc" -eq 0 ] && has "::warning title=kibi query time::kibi coverage took 90 s" && clean_stdout && echo yes || echo no)"

timed 95 "" && rc=0 || rc=$?
check "a slow query outside Actions warns in plain text" \
  "$([ "$rc" -eq 0 ] && has "warning: kibi coverage took 95 s" && lacks "::warning" && echo yes || echo no)"

timed 10 "" && rc=0 || rc=$?
check "kibi's stderr and arguments pass through" "$(has "kibi on stderr: coverage --format json" && echo yes || echo no)"

FAKE_STATUS=3 timed 90 true && rc=0 || rc=$?
check "kibi's exit status passes through, with the duration and the warning still printed" \
  "$([ "$rc" -eq 3 ] && has "took 90 s" && has "::warning" && echo yes || echo no)"

bash "$root/scripts/kibi-timed.sh" "$work/kibi" >"$work/out" 2>"$work/err" && rc=0 || rc=$?
check "a missing kibi command is a usage error" "$([ "$rc" -eq 2 ] && has usage && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-timed.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-timed.test.sh: all $cases checks passed"
