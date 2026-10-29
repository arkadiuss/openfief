#!/usr/bin/env bash
# Tests for scripts/ci-images.sh in a throwaway directory (kibi: TEST-tooling-ci-images).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
failures=0

expect_output() {
  local name="$1" want="$2" got
  if ! got="$(bash "$root/scripts/ci-images.sh" ghcr.io/Example/OpenFief 2>&1)"; then
    echo "FAIL: $name should pass, got: $got" >&2
    failures=$((failures + 1))
  elif [ "$got" != "$want" ]; then
    echo "FAIL: $name: want $want, got $got" >&2
    failures=$((failures + 1))
  fi
}

expect_fail() {
  local name="$1" pattern="$2" out
  if out="$(bash "$root/scripts/ci-images.sh" 2>&1)"; then
    echo "FAIL: $name should fail" >&2
    failures=$((failures + 1))
  elif ! grep -q "$pattern" <<<"$out"; then
    echo "FAIL: $name should report '$pattern', got: $out" >&2
    failures=$((failures + 1))
  fi
}

mkdir -p cmd/no-image
expect_output "no Dockerfile" '[]'

mkdir -p cmd/fief-operator plugins/xtb-import
touch cmd/fief-operator/Dockerfile plugins/xtb-import/Dockerfile
expect_output "two images" '[{"name":"fief-operator","dockerfile":"cmd/fief-operator/Dockerfile","image":"ghcr.io/example/openfief/fief-operator"},{"name":"xtb-import","dockerfile":"plugins/xtb-import/Dockerfile","image":"ghcr.io/example/openfief/xtb-import"}]'

expect_kind() {
  local kind="$1" want="$2" got
  if ! got="$(bash "$root/scripts/ci-images.sh" ghcr.io/example/openfief "$kind" 2>&1)" || [ "$got" != "$want" ]; then
    echo "FAIL: kind $kind: want $want, got $got" >&2
    failures=$((failures + 1))
  fi
}
expect_kind plugins '[{"name":"xtb-import","dockerfile":"plugins/xtb-import/Dockerfile","image":"ghcr.io/example/openfief/xtb-import"}]'
expect_kind commands '[{"name":"fief-operator","dockerfile":"cmd/fief-operator/Dockerfile","image":"ghcr.io/example/openfief/fief-operator"}]'
if out="$(bash "$root/scripts/ci-images.sh" ghcr.io/example/openfief images 2>&1)" || ! grep -q "must be all, plugins or commands" <<<"$out"; then
  echo "FAIL: an unknown kind should fail, got: $out" >&2
  failures=$((failures + 1))
fi

mkdir -p plugins/fief-operator
touch plugins/fief-operator/Dockerfile
expect_fail "a duplicate name" "'fief-operator' is used by another Dockerfile"
rm -r plugins/fief-operator

mkdir -p cmd/Bad_Name
touch cmd/Bad_Name/Dockerfile
expect_fail "an invalid name" "image name 'Bad_Name' must be lowercase"

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "ci-images: 7 cases passed"
