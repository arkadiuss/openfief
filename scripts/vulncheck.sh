#!/usr/bin/env bash
# Vulnerability checks for Go (govulncheck) and the npm dev tooling (npm audit).
# Any failure fails the check: a finding, and also an unreachable vulnerability database.
set -uo pipefail

status=0

if [ -f go.mod ]; then
  if ! out="$(govulncheck ./... 2>&1)"; then
    echo "vulncheck: govulncheck failed:" >&2
    echo "$out" >&2
    status=1
  fi
else
  echo "vulncheck: no Go module yet, skipping govulncheck"
fi

if ! out="$(npm audit --audit-level=high 2>&1)"; then
  echo "vulncheck: npm audit failed:" >&2
  echo "$out" >&2
  status=1
fi

if [ "$status" -eq 0 ]; then
  echo "vulncheck: no known vulnerabilities"
fi
exit "$status"
