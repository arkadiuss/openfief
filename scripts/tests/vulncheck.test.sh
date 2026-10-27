#!/usr/bin/env bash
# Tests for scripts/vulncheck.sh with an unreachable npm registry (kibi: TEST-tooling-vuln-offline).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# A copy with only the npm lockfile: no go.mod, so govulncheck never runs and the test stays offline.
mkdir -p "$work/scripts"
cp "$root/package.json" "$root/package-lock.json" "$work/"
cp "$root/scripts/vulncheck.sh" "$work/scripts/"
cd "$work"
failures=0

# Port 9 (discard) on localhost refuses connections, which simulates being offline.
export npm_config_registry="http://127.0.0.1:9/"

if ! out="$(env -u CI bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: offline outside CI should warn and pass, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'unreachable, skipped locally' <<<"$out"; then
  echo "FAIL: offline outside CI should print a warning, got: $out" >&2
  failures=$((failures + 1))
fi

if out="$(CI=true bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: offline in CI should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'unreachable in CI' <<<"$out"; then
  echo "FAIL: offline in CI should say why, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "vulncheck: 2 offline cases passed"
