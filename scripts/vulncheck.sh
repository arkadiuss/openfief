#!/usr/bin/env bash
# Vulnerability checks for Go (govulncheck) and the npm dev tooling (npm audit).
# Any failure fails the check: a finding, and also an unreachable vulnerability database.
# Arguments are the Go module directories to check (default: the root module).
set -uo pipefail

status=0

if [ "$#" -eq 0 ]; then
  set -- .
fi
for module in "$@"; do
  if [ ! -f "$module/go.mod" ]; then
    echo "vulncheck: no Go module in $module, skipping govulncheck"
    continue
  fi
  if ! out="$(cd "$module" && govulncheck ./... 2>&1)"; then
    echo "vulncheck: govulncheck failed in $module:" >&2
    echo "$out" >&2
    status=1
  fi
done

# npm audit: every high or critical advisory fails the check, except one listed in scripts/npm-audit-allow.txt before
# its expiry; an entry past its expiry fails it too. A report that is no JSON (no network, no registry) fails it.
allow="$(dirname "$0")/npm-audit-allow.txt"
today="$(date -u +%Y-%m-%d)"
report="$(npm audit --json 2>/dev/null)"
if ! found="$(gojq -r '[.vulnerabilities[]?.via[]? | objects | select(.severity == "high" or .severity == "critical")
    | .url | sub("^.*/"; "")] | unique[]' <<<"$report" 2>/dev/null)" || ! gojq -e '.vulnerabilities' <<<"$report" >/dev/null 2>&1; then
  echo "vulncheck: npm audit failed: no audit report (is the registry reachable?)" >&2
  npm audit --audit-level=high >&2 2>&1 || true
  status=1
else
  blocking=""
  for id in $found; do
    expiry="$( [ -f "$allow" ] && awk -v id="$id" '$1 == id {print $2}' "$allow" | head -1)"
    if [ -n "$expiry" ] && [[ ! "$today" > "$expiry" ]]; then
      echo "vulncheck: npm audit: $id allowed until $expiry (scripts/npm-audit-allow.txt)"
    elif [ -n "$expiry" ]; then
      echo "vulncheck: npm audit: $id was allowed until $expiry, which has passed (scripts/npm-audit-allow.txt)" >&2
      blocking="$blocking $id"
    else
      blocking="$blocking $id"
    fi
  done
  if [ -n "$blocking" ]; then
    echo "vulncheck: npm audit failed:$blocking" >&2
    npm audit --audit-level=high >&2 2>&1 || true
    status=1
  fi
fi

if [ "$status" -eq 0 ]; then
  echo "vulncheck: no known vulnerabilities outside scripts/npm-audit-allow.txt"
fi
exit "$status"
