#!/usr/bin/env bash
# Runs a command that downloads something (curl, go install, npm ci, gh release download) and runs it again when
# it failed for a reason a later try can fix: the server answered 429 or 5xx, the connection failed, reset or
# timed out, or a name didn't resolve. Any other failure, a 4xx answer other than 429 above all, is final at once:
# retrying a 404 or a failed checksum only hides it for a minute (docs/dev/pipeline.md#download-retries).
# The reason is read from the command's own error text, so the command must say what went wrong: use curl -sS (not
# -s alone), and keep the checksum check after the download, exactly as strict as before, in the caller.
#
# The first try is followed by up to five more, after 2, 4, 8, 16 and 32 seconds: a run waits at most 62 seconds
# for a server that comes back. RETRY_ATTEMPTS (default 6), RETRY_DELAY (the first wait, default 2) and
# RETRY_MAX_DELAY (default 32) change that; RETRY_SLEEP names the command that waits (tests pass a fake). Each wait
# and its reason are printed on stderr; standard output and a successful command's output are untouched.
# Usage: retry.sh [--] <command> [arguments...]
set -uo pipefail

if [ "${1:-}" = "--" ]; then shift; fi
if [ "$#" -eq 0 ]; then
  echo "usage: retry.sh [--] <command> [arguments...]" >&2
  exit 2
fi

attempts="${RETRY_ATTEMPTS:-6}"
delay="${RETRY_DELAY:-2}"
max_delay="${RETRY_MAX_DELAY:-32}"
sleeper="${RETRY_SLEEP:-sleep}"

# Error text of a failure worth another try. Only 429 and 5xx count among HTTP answers; the other 4xx never match.
transient='curl: \((6|7|18|28|35|52|55|56|92)\)'
transient+='|returned error: (429|5[0-9][0-9])'
transient+='|(^|[^0-9])(429|50[0-9]) (Too Many Requests|Internal Server Error|Not Implemented|Bad Gateway|Service Unavailable|Gateway Time-?out|Gateway Timeout)'
transient+='|HTTP (429|5[0-9][0-9])'
transient+='|i/o timeout|TLS handshake timeout|connection reset|connection refused|unexpected EOF|server misbehaving'
transient+='|no such host|Temporary failure in name resolution|Operation timed out|ETIMEDOUT|ECONNRESET|ECONNREFUSED'
transient+='|EAI_AGAIN|ESOCKETTIMEDOUT|socket hang up'

err="$(mktemp)"
trap 'rm -f "$err"' EXIT

attempt=1
while true; do
  "$@" 2>"$err"
  status=$?
  cat "$err" >&2
  if [ "$status" -eq 0 ]; then
    exit 0
  fi
  if ! grep -qiE "$transient" "$err"; then
    exit "$status"
  fi
  if [ "$attempt" -ge "$attempts" ]; then
    echo "retry: $1 failed $attempts times in a row with a transient error; giving up" >&2
    exit "$status"
  fi
  reason="$(grep -iE -m1 "$transient" "$err" | cut -c1-160)"
  echo "retry: try $attempt of $attempts failed ($reason); trying again in ${delay}s" >&2
  "$sleeper" "$delay"
  delay=$((delay * 2))
  if [ "$delay" -gt "$max_delay" ]; then delay="$max_delay"; fi
  attempt=$((attempt + 1))
done
