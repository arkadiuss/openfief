#!/usr/bin/env bash
# Tests for scripts/demo/validate.sh on copies of deploy/ (kibi: TEST-tooling-demo-validate).
# Needs kustomize and gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# fresh copies deploy/ into a throwaway repository root, so Flux paths (./deploy/...) resolve the same way.
fresh() {
  rm -rf "$work/repo"
  mkdir -p "$work/repo"
  cp -r "$root/deploy" "$work/repo/deploy"
}

# validate runs validate.sh in the throwaway root; output lands in $work/out.
validate() {
  (cd "$work/repo" && bash "$root/scripts/demo/validate.sh") >"$work/out" 2>&1
}

# expect records whether validate passed or failed as wanted, and that the output names the problem.
expect() {
  local name="$1" want="$2" pattern="$3" got=pass
  cases=$((cases + 1))
  validate || got=fail
  if [ "$got" != "$want" ] || ! grep -q "$pattern" "$work/out"; then
    echo "FAIL: $name: want $want with '$pattern', got $got:" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

fresh
expect "the committed manifests" pass "every image is pinned by digest"

fresh
sed -i 's/    provider: cosign/    provider: notation/' "$work/repo/deploy/demo/stack.yaml"
expect "an OCI source without cosign verification" fail "OCI sources without cosign verification: openfief-demo-app"

fresh
sed -i '/^images:/,$d' "$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "an image pinned by tag only" fail "not pinned by digest: ghcr.io/arkadiuss/openfief/"

fresh
sed -i 's#path: ./deploy/demo/database#path: ./deploy/demo/missing#' "$work/repo/deploy/demo/stack.yaml"
expect "a Flux path that does not build" fail "kustomize build ./deploy/demo/missing failed"

fresh
sed -i 's/[$]{DEMO_HOST}/${DEMO_HOSTNAME}/' "$work/repo/deploy/demo/routes/core.yaml"
expect "a variable the demo settings don't define" fail "uses variables the demo settings don't define: DEMO_HOSTNAME"

if [ "$failures" -ne 0 ]; then
  echo "demo-validate: $failures of $cases cases failed" >&2
  exit 1
fi
echo "demo-validate: $cases cases passed"
