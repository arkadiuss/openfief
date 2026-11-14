#!/usr/bin/env bash
# Tests for scripts/kibi-install-status.sh on a temporary project tree (a package-lock.json, node_modules with fake
# package.json files and a patches directory): no network, no npm, the real node_modules untouched
# (kibi: TEST-tooling-kibi-install-tracks-pins).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# fresh builds the project in $work/p: kibi-cli 2.42.0 and kibi-core 0.31.0 pinned and installed, one patch for each,
# a stamp from the script, and a non-kibi package whose version the script must ignore.
fresh() {
  rm -rf "$work/p"
  mkdir -p "$work/p/node_modules/kibi-cli" "$work/p/node_modules/kibi-core" "$work/p/node_modules/mermaid" "$work/p/patches"
  cat >"$work/p/package-lock.json" <<'EOF'
{"lockfileVersion": 3, "packages": {
  "": {"name": "x"},
  "node_modules/kibi-cli": {"version": "2.42.0"},
  "node_modules/kibi-core": {"version": "0.31.0"},
  "node_modules/kibi-cli/node_modules/kibi-core": {"version": "0.1.0"},
  "node_modules/kibi-swipl-other-os": {"version": "1.0.2", "optional": true, "os": ["darwin"]},
  "node_modules/mermaid": {"version": "12.2.0"}
}}
EOF
  installed kibi-cli 2.42.0
  installed kibi-core 0.31.0
  installed mermaid 11.0.0
  echo 'patch one' >"$work/p/patches/kibi-cli+2.42.0.patch"
  echo 'patch two' >"$work/p/patches/kibi-core+0.31.0.patch"
  bash "$root/scripts/kibi-install-status.sh" stamp "$work/p"
}

# installed <package> <version> writes the package.json npm would.
installed() { echo "{\"name\": \"$1\", \"version\": \"$2\"}" >"$work/p/node_modules/$1/package.json"; }

# status runs the check on $work/p, keeping its output.
status() { bash "$root/scripts/kibi-install-status.sh" check "$work/p" >"$work/out" 2>"$work/err"; }

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  err | /' "$work/err" >&2
    failures=$((failures + 1))
  fi
}

# says is true when the last run's stderr holds the fixed string.
says() { grep -qF -- "$1" "$work/err"; }

fresh
status && rc=0 || rc=$?
check "a stamped install at the pinned versions is up to date, quietly" \
  "$([ "$rc" -eq 0 ] && [ ! -s "$work/err" ] && echo yes || echo no)"

installed kibi-cli 2.36.0
status && rc=0 || rc=$?
check "an older installed kibi-cli than the lock pins fails and names both versions" \
  "$([ "$rc" -eq 1 ] && says "kibi-cli 2.42.0" && says "kibi-cli 2.36.0" && echo yes || echo no)"

fresh
sed -i 's/"version": "0.31.0"}/"version": "0.32.0"}/' "$work/p/package-lock.json"
status && rc=0 || rc=$?
check "a lock that moved on past the installed kibi-core fails" \
  "$([ "$rc" -eq 1 ] && says "kibi-core 0.32.0" && says "kibi-core 0.31.0" && echo yes || echo no)"

fresh
rm "$work/p/node_modules/kibi-core/package.json"
status && rc=0 || rc=$?
check "a pinned package that isn't installed fails" "$([ "$rc" -eq 1 ] && says "kibi-core missing" && echo yes || echo no)"

fresh
status && rc=0 || rc=$?
check "an optional package npm skipped on this platform is not missing" "$([ "$rc" -eq 0 ] && echo yes || echo no)"

fresh
mkdir "$work/p/node_modules/kibi-swipl-other-os"
installed kibi-swipl-other-os 0.9.0
status && rc=0 || rc=$?
check "an optional package that is installed must match its pin" \
  "$([ "$rc" -eq 1 ] && says "kibi-swipl-other-os 1.0.2" && echo yes || echo no)"

fresh
installed mermaid 1.0.0
status && rc=0 || rc=$?
check "a package that isn't kibi's is not looked at" "$([ "$rc" -eq 0 ] && echo yes || echo no)"

fresh
echo 'patch two, changed' >"$work/p/patches/kibi-core+0.31.0.patch"
status && rc=0 || rc=$?
check "a patch edited since the install fails and names it" \
  "$([ "$rc" -eq 1 ] && says "kibi-core+0.31.0.patch" && echo yes || echo no)"

fresh
echo 'patch three' >"$work/p/patches/kibi-lang-go+1.0.0.patch"
status && rc=0 || rc=$?
check "a patch added since the install fails" "$([ "$rc" -eq 1 ] && says "kibi-lang-go+1.0.0.patch" && echo yes || echo no)"

fresh
rm "$work/p/patches/kibi-cli+2.42.0.patch"
status && rc=0 || rc=$?
check "a patch removed since the install fails" "$([ "$rc" -eq 1 ] && says "kibi-cli+2.42.0.patch" && echo yes || echo no)"

fresh
rm "$work/p/node_modules/.kibi-install-stamp"
status && rc=0 || rc=$?
check "an install with no stamp fails: its patches are not known to be applied" \
  "$([ "$rc" -eq 1 ] && says "no install stamp" && echo yes || echo no)"

fresh
installed kibi-cli 2.42.0
echo '{"name": "kibi-cli", "version": "2.42.0", "reinstalled": true}' >"$work/p/node_modules/kibi-cli/package.json"
status && rc=0 || rc=$?
check "a package.json that changed without its version changing is still up to date" \
  "$([ "$rc" -eq 0 ] && echo yes || echo no)"

fresh
rm -r "$work/p/patches"
mkdir "$work/p/patches"
rm "$work/p/node_modules/.kibi-install-stamp"
bash "$root/scripts/kibi-install-status.sh" stamp "$work/p" >"$work/out" 2>"$work/err" && rc=0 || rc=$?
status && rc2=0 || rc2=$?
check "a project with no patches stamps and is then up to date" "$([ "$rc" -eq 0 ] && [ "$rc2" -eq 0 ] && echo yes || echo no)"

fresh
installed kibi-cli 2.36.0
rm "$work/p/node_modules/.kibi-install-stamp"
bash "$root/scripts/kibi-install-status.sh" stamp "$work/p" >"$work/out" 2>"$work/err" && rc=0 || rc=$?
check "stamp refuses to vouch for an install that differs from the pins" \
  "$([ "$rc" -eq 1 ] && [ ! -e "$work/p/node_modules/.kibi-install-stamp" ] && says "not stamping" && echo yes || echo no)"

bash "$root/scripts/kibi-install-status.sh" frobnicate "$work/p" >"$work/out" 2>"$work/err" && rc=0 || rc=$?
check "an unknown mode is a usage error" "$([ "$rc" -eq 2 ] && says usage && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "kibi-install-status.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "kibi-install-status.test.sh: all $cases checks passed"
