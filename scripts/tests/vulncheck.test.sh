#!/usr/bin/env bash
# Tests for scripts/vulncheck.sh with unreachable vulnerability databases (kibi: TEST-tooling-vuln-offline).
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

# A go.mod plus a stub govulncheck that fails like a proxy refusing vuln.go.dev: no network needed.
mkdir -p "$work/bin"
cat >"$work/bin/govulncheck" <<'STUB'
#!/usr/bin/env bash
echo 'govulncheck: fetching vulnerabilities: Get "https://vuln.go.dev/index/modules.json.gz": Forbidden' >&2
exit 1
STUB
chmod +x "$work/bin/govulncheck"
echo 'module example.com/offline' >"$work/go.mod"

if ! out="$(env -u CI PATH="$work/bin:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: refused vuln.go.dev outside CI should warn and pass, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'vuln.go.dev unreachable, skipped locally' <<<"$out"; then
  echo "FAIL: refused vuln.go.dev outside CI should print a warning, got: $out" >&2
  failures=$((failures + 1))
fi

if out="$(CI=true PATH="$work/bin:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: refused vuln.go.dev in CI should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'vuln.go.dev unreachable in CI' <<<"$out"; then
  echo "FAIL: refused vuln.go.dev in CI should say why, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "vulncheck: 4 offline cases passed"
