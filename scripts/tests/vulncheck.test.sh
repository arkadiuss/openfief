#!/usr/bin/env bash
# Tests for scripts/vulncheck.sh: a finding and an unreachable registry both fail it (kibi: TEST-tooling-vuln-offline).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# A copy with only the npm lockfile: no go.mod, so govulncheck never runs.
mkdir -p "$work/scripts" "$work/stub"
cp "$root/package.json" "$root/package-lock.json" "$work/"
cp "$root/scripts/vulncheck.sh" "$work/scripts/"
cd "$work"
failures=0

# A stub npm whose audit reports a high-severity finding, so the case needs no network.
cat >stub/npm <<'EOF'
#!/usr/bin/env bash
echo "example-package  <1.2.3"
echo "Severity: high"
echo "1 high severity vulnerability"
exit 1
EOF
chmod +x stub/npm

if out="$(PATH="$work/stub:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: a reported vulnerability should fail the check, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q '1 high severity vulnerability' <<<"$out"; then
  echo "FAIL: a reported vulnerability should be printed, got: $out" >&2
  failures=$((failures + 1))
fi

# Port 9 (discard) on localhost refuses connections, so the advisory database is unreachable.
if out="$(npm_config_registry="http://127.0.0.1:9/" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: an unreachable registry should fail the check, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'npm audit failed' <<<"$out"; then
  echo "FAIL: an unreachable registry should be reported, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "vulncheck: 2 cases passed"
