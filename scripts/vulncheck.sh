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

if ! out="$(npm audit --audit-level=high 2>&1)"; then
  echo "vulncheck: npm audit failed:" >&2
  echo "$out" >&2
  status=1
fi

if [ "$status" -eq 0 ]; then
  echo "vulncheck: no known vulnerabilities"
fi
exit "$status"
