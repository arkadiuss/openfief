#!/usr/bin/env bash
# Tests for scripts/vulncheck.sh: a finding or an unreachable database fails it (kibi: TEST-tooling-vuln-offline).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# A copy with only the npm lockfile: no go.mod yet, so govulncheck doesn't run in the first two cases.
mkdir -p "$work/scripts" "$work/stub-finding" "$work/stub-clean" "$work/bin"
cp "$root/package.json" "$root/package-lock.json" "$work/"
cp "$root/scripts/vulncheck.sh" "$work/scripts/"
cd "$work"
failures=0

# A stub npm whose audit reports one high-severity advisory (JSON with --json, text without), so the cases need no
# network. The advisory's id comes from ADVISORY, GHSA-aaaa-bbbb-cccc by default.
cat >stub-finding/npm <<'EOF'
#!/usr/bin/env bash
id="${ADVISORY:-GHSA-aaaa-bbbb-cccc}"
if [ "${2:-}" = --json ]; then
  printf '{"vulnerabilities":{"example-package":{"severity":"high","via":[{"severity":"high","url":"https://github.com/advisories/%s"}]}}}\n' "$id"
else
  echo "example-package  <1.2.3"
  echo "Severity: high"
  echo "1 high severity vulnerability"
fi
exit 1
EOF
chmod +x stub-finding/npm

if out="$(PATH="$work/stub-finding:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
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

# The allowlist (scripts/npm-audit-allow.txt): an advisory listed before its expiry passes and is named; one past its
# expiry fails; one not listed still fails.
printf 'GHSA-aaaa-bbbb-cccc 2999-12-31 a test entry\nGHSA-dddd-eeee-ffff 2000-01-01 an expired test entry\n' >scripts/npm-audit-allow.txt
if ! out="$(PATH="$work/stub-finding:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: an advisory allowed until a later day should pass, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'GHSA-aaaa-bbbb-cccc allowed until 2999-12-31' <<<"$out"; then
  echo "FAIL: an allowed advisory should be named with its expiry, got: $out" >&2
  failures=$((failures + 1))
fi
if out="$(ADVISORY=GHSA-dddd-eeee-ffff PATH="$work/stub-finding:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: an advisory whose allowance expired should fail, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'which has passed' <<<"$out"; then
  echo "FAIL: an expired allowance should be reported, got: $out" >&2
  failures=$((failures + 1))
fi
if out="$(ADVISORY=GHSA-zzzz-yyyy-xxxx PATH="$work/stub-finding:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: an advisory not on the allowlist should fail, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'npm audit failed: GHSA-zzzz-yyyy-xxxx' <<<"$out"; then
  echo "FAIL: an advisory not on the allowlist should be named, got: $out" >&2
  failures=$((failures + 1))
fi
rm -f scripts/npm-audit-allow.txt

# A go.mod plus a stub govulncheck that fails like a proxy refusing vuln.go.dev, and a clean npm
# audit, so only the Go database is unreachable. No network needed.
cat >bin/govulncheck <<'EOF'
#!/usr/bin/env bash
echo 'govulncheck: fetching vulnerabilities: Get "https://vuln.go.dev/index/modules.json.gz": Forbidden' >&2
exit 1
EOF
cat >stub-clean/npm <<'EOF'
#!/usr/bin/env bash
if [ "${2:-}" = --json ]; then echo '{"vulnerabilities":{}}'; else echo "found 0 vulnerabilities"; fi
EOF
chmod +x bin/govulncheck stub-clean/npm
echo 'module example.com/offline' >go.mod

if out="$(PATH="$work/bin:$work/stub-clean:$PATH" bash scripts/vulncheck.sh 2>&1)"; then
  echo "FAIL: an unreachable vuln.go.dev should fail the check, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'govulncheck failed' <<<"$out"; then
  echo "FAIL: an unreachable vuln.go.dev should be reported, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "vulncheck: 6 cases passed"
