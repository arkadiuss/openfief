#!/usr/bin/env bash
# Vulnerability checks for Go (govulncheck) and the npm dev tooling (npm audit).
# Findings always fail. An unreachable database warns locally and fails in CI (CI=true).
set -uo pipefail

status=0

unreachable() {
  if [ "${CI:-}" = "true" ]; then
    echo "vulncheck: $1 unreachable in CI" >&2
    status=1
  else
    echo "vulncheck: warning: $1 unreachable, skipped locally (CI enforces it)" >&2
  fi
}

network_error='dial|lookup|no such host|connection refused|i/o timeout|ENOTFOUND|ECONNREFUSED|ETIMEDOUT|EAI_AGAIN'

if [ -f go.mod ]; then
  out="$(govulncheck ./... 2>&1)"
  rc=$?
  if [ "$rc" -ne 0 ]; then
    if [ "$rc" -ne 3 ] && grep -qiE "$network_error" <<<"$out"; then
      unreachable "vuln.go.dev"
    else
      echo "$out" >&2
      status=1
    fi
  fi
else
  echo "vulncheck: no Go module yet, skipping govulncheck"
fi

out="$(npm audit --audit-level=high 2>&1)"
rc=$?
if [ "$rc" -ne 0 ]; then
  if grep -qiE "$network_error" <<<"$out"; then
    unreachable "the npm registry"
  else
    echo "$out" >&2
    status=1
  fi
fi

if [ "$status" -eq 0 ]; then
  echo "vulncheck: no known vulnerabilities"
fi
exit "$status"
