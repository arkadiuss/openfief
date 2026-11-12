#!/usr/bin/env bash
# Tests for scripts/retry.sh, which runs a download again when the failure is one a later try can fix, with a fake
# flaky command and a fake sleep that records the waits (kibi: TEST-tooling-retry).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# The fake command fails with each line of $FLAKY_SCRIPT in turn (an exit status, a space, the error text), then
# succeeds; it counts its calls in $work/calls and prints its arguments, one per line, on standard output when it
# succeeds.
cat >"$work/flaky" <<'EOF'
#!/usr/bin/env bash
n=$(($(cat "$FLAKY_CALLS" 2>/dev/null || echo 0) + 1))
echo "$n" >"$FLAKY_CALLS"
line="$(sed -n "${n}p" "$FLAKY_SCRIPT")"
if [ -n "$line" ]; then
  echo "${line#* }" >&2
  exit "${line%% *}"
fi
printf '%s\n' "$@"
EOF
cat >"$work/sleep" <<'EOF'
#!/usr/bin/env bash
echo "$1" >>"$FLAKY_WAITS"
EOF
chmod +x "$work/flaky" "$work/sleep"

# run executes retry.sh on the fake command after the failures in $1 (lines of "status text"), with the other
# arguments after it, and any extra environment first in the RETRY_ENV array; it leaves the exit status in $rc.
run() {
  local script="$1"
  shift
  printf '%s' "$script" >"$work/script"
  rm -f "$work/calls" "$work/waits"
  : >"$work/waits"
  rc=0
  env FLAKY_SCRIPT="$work/script" FLAKY_CALLS="$work/calls" FLAKY_WAITS="$work/waits" RETRY_SLEEP="$work/sleep" \
    "${RETRY_ENV[@]}" bash "$root/scripts/retry.sh" "$@" >"$work/out" 2>"$work/err" || rc=$?
}
RETRY_ENV=(RETRY_DELAY=2)
calls() { cat "$work/calls"; }
waits() { tr '\n' ' ' <"$work/waits" | sed 's/ $//'; }

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name (exit $rc, $(calls) calls, waits: $(waits))" >&2
    sed 's/^/  out | /' "$work/out" >&2
    sed 's/^/  err | /' "$work/err" >&2
    failures=$((failures + 1))
  fi
}
is() { if [ "$1" = "$2" ]; then echo yes; else echo no; fi; }

run '' "$work/flaky" a
check "a command that works runs once, without waiting" "$(is "$rc/$(calls)/$(waits)" '0/1/')"

run '' -- "$work/flaky" "an argument with spaces" b
check "-- ends retry's own arguments, and every argument arrives whole" \
  "$(is "$(cat "$work/out")" "$(printf 'an argument with spaces\nb')")"

run $'22 curl: (22) The requested URL returned error: 504\n22 curl: (22) The requested URL returned error: 504\n' "$work/flaky"
check "a 504 twice is waited out with doubling waits, and then the download works" "$(is "$rc/$(calls)/$(waits)" '0/3/2 4')"
check "each wait names its reason" "$(if grep -q 'try 1 of 6 failed (curl: (22) The requested URL returned error: 504); trying again in 2s' "$work/err"; then echo yes; else echo no; fi)"

for answer in 500 502 503; do
  run "22 curl: (22) The requested URL returned error: $answer"$'\n' "$work/flaky"
  check "a $answer is retried" "$(is "$rc/$(calls)" '0/2')"
done

run $'22 curl: (22) The requested URL returned error: 429\n' "$work/flaky"
check "a 429 is retried" "$(is "$rc/$(calls)" '0/2')"

for answer in 400 401 403 404 410; do
  run "22 curl: (22) The requested URL returned error: $answer"$'\n' "$work/flaky"
  check "a $answer is final at once, with curl's exit status" "$(is "$rc/$(calls)/$(waits)" '22/1/')"
done

for code in 6 7 28 35 52 55 56; do
  run "$code curl: ($code) Some network failure"$'\n' "$work/flaky"
  check "curl error $code (a connection that didn't work) is retried" "$(is "$rc/$(calls)" '0/2')"
done

run $'60 curl: (60) SSL certificate problem: unable to get local issuer certificate\n' "$work/flaky"
check "a certificate problem is final at once" "$(is "$rc/$(calls)" '60/1')"

run $'1 go: golang.org/x/vuln@v1.9.0: reading https://proxy.golang.org/x/@v/v1.9.0.zip: 503 Service Unavailable\n' "$work/flaky"
check "go's proxy answering 503 is retried" "$(is "$rc/$(calls)" '0/2')"
run $'1 dial tcp 192.0.2.1:443: i/o timeout\n' "$work/flaky"
check "a timeout is retried" "$(is "$rc/$(calls)" '0/2')"
run $'1 npm error code ECONNRESET\n' "$work/flaky"
check "npm's connection reset is retried" "$(is "$rc/$(calls)" '0/2')"
run $'1 HTTP 502: Bad Gateway (https://api.github.com/repos/x/y/releases)\n' "$work/flaky"
check "gh's HTTP 502 is retried" "$(is "$rc/$(calls)" '0/2')"
run $'1 Error: signing [x@sha256:00]: signing digest: getting keypair and token: retrieving ID token: reading ID token: fetching ambient OIDC credentials: invalid character \'u\' looking for beginning of value\n' "$work/flaky"
check "cosign reading an error page from GitHub's token service is retried" "$(is "$rc/$(calls)" '0/2')"
run $'1 Error: signing [x@sha256:00]: getting key from Fulcio: retrieving cert: POST https://fulcio.sigstore.dev/api/v2/signingCert: 403 Forbidden\n' "$work/flaky"
check "cosign refused by Fulcio is final at once" "$(is "$rc/$(calls)" '1/1')"

run $'2 main.go:3:1: syntax error: unexpected }\n' "$work/flaky"
check "a failure that isn't the network's is final at once, with its exit status" "$(is "$rc/$(calls)/$(waits)" '2/1/')"
check "and its error text still reaches the caller" "$(if grep -q 'syntax error' "$work/err"; then echo yes; else echo no; fi)"

always="$(printf '22 curl: (22) The requested URL returned error: 504\n%.0s' 1 2 3 4 5 6 7 8 9 10 11 12)"$'\n'
run "$always" "$work/flaky"
check "a server that never comes back is tried six times, waiting 2, 4, 8, 16, 32 seconds, and then fails" \
  "$(is "$rc/$(calls)/$(waits)" '22/6/2 4 8 16 32')"
check "giving up says so" "$(if grep -q 'retry: .*flaky failed 6 times in a row' "$work/err"; then echo yes; else echo no; fi)"

RETRY_ENV=(RETRY_DELAY=2 RETRY_ATTEMPTS=9)
run "$always" "$work/flaky"
check "the wait stops growing at 32 seconds" "$(is "$rc/$(calls)/$(waits)" '22/9/2 4 8 16 32 32 32 32')"

RETRY_ENV=(RETRY_DELAY=1 RETRY_ATTEMPTS=2 RETRY_MAX_DELAY=5)
run "$always" "$work/flaky"
check "RETRY_ATTEMPTS and RETRY_DELAY are honoured" "$(is "$rc/$(calls)/$(waits)" '22/2/1')"
RETRY_ENV=(RETRY_DELAY=2)

rc=0
bash "$root/scripts/retry.sh" >/dev/null 2>&1 || rc=$?
check "no command is a usage error" "$(is "$rc" 2)"

if [ "$failures" -gt 0 ]; then
  echo "retry.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "retry.test.sh: all $cases checks passed"
