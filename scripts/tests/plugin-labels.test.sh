#!/usr/bin/env bash
# Tests for scripts/plugin-labels.sh on copies of plugins/ and deploy/demo/ (kibi: TEST-tooling-plugin-labels).
# Needs gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# fresh copies the plugins and the demo into a throwaway repository root.
fresh() {
  rm -rf "$work/repo"
  mkdir -p "$work/repo/deploy"
  cp -r "$root/plugins" "$work/repo/plugins"
  if [ -d "$root/deploy/demo" ]; then
    cp -r "$root/deploy/demo" "$work/repo/deploy/demo"
  fi
}

# run runs plugin-labels.sh in mode on the throwaway root; output lands in $work/out.
run() {
  bash "$root/scripts/plugin-labels.sh" "$1" "$work/repo" >"$work/out" 2>&1
}

# expect records whether a run passed or failed as wanted, and that the output names the problem.
expect() {
  local name="$1" mode="$2" want="$3" pattern="$4" got=pass
  cases=$((cases + 1))
  run "$mode" || got=fail
  if [ "$got" != "$want" ] || ! grep -q -- "$pattern" "$work/out"; then
    echo "FAIL: $name: want $want with '$pattern', got $got:" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

fresh
expect "the committed labels" check pass "every plugin label matches its manifest"

fresh
gojq '.egress += ["extra.example.com"]' "$root/plugins/nbp-fx/manifest.json" >"$work/repo/plugins/nbp-fx/manifest.json"
expect "a manifest changed without its label" check fail "plugins/nbp-fx/Dockerfile's io.openfief.plugin.manifest label differs"
expect "write brings the label in step" write pass "every plugin label matches"
cases=$((cases + 1))
if ! grep -q 'extra.example.com' "$work/repo/plugins/nbp-fx/Dockerfile"; then
  echo "FAIL: write did not put the new host in the label" >&2
  failures=$((failures + 1))
fi

fresh
gojq '.configSchema.description = "it'"'"'s"' "$root/plugins/nbp-fx/manifest.json" >"$work/repo/plugins/nbp-fx/manifest.json"
expect "a single quote the label can't hold" write fail "contains a single quote"

fresh
gojq '.configSchema.properties.start.pattern = "^[0-9]{4}\\\\d$"' "$root/plugins/nbp-fx/manifest.json" >"$work/repo/plugins/nbp-fx/manifest.json"
sed -i '/^  - nbp.yaml$/d' "$work/repo/deploy/demo/plugins/kustomization.yaml" 2>/dev/null || true
rm -f "$work/repo/deploy/demo/plugins/nbp.yaml"
expect "backslashes are written as they are" write pass "every plugin label matches"
expect "and read back the same" check pass "every plugin label matches"

fresh
sed -i '/^LABEL io.openfief.plugin.manifest=/d' "$work/repo/plugins/ibkr-flex/Dockerfile"
expect "a Dockerfile without the label line" write fail "plugins/ibkr-flex/Dockerfile has no LABEL io.openfief.plugin.manifest line"

expect "an unknown mode" verify fail "usage: plugin-labels.sh check|write"

if [ "$failures" -ne 0 ]; then
  echo "plugin-labels: $failures of $cases cases failed" >&2
  exit 1
fi
echo "plugin-labels: $cases cases passed"
