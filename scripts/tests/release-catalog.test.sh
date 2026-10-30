#!/usr/bin/env bash
# Tests for scripts/release-catalog.sh in a throwaway directory (kibi: TEST-tooling-release-catalog). Needs gojq on PATH.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
failures=0
# Synthetic digests; not of any real image.
da="sha256:$(printf 'a%.0s' {1..64})"
db="sha256:$(printf 'b%.0s' {1..64})"

run() { bash "$root/scripts/release-catalog.sh" "$@" 2>&1; }

expect_fail() {
  local name="$1" pattern="$2" out
  shift 2
  if out="$(run "$@")"; then
    echo "FAIL: $name should fail" >&2
    failures=$((failures + 1))
  elif ! grep -q "$pattern" <<<"$out"; then
    echo "FAIL: $name should report '$pattern', got: $out" >&2
    failures=$((failures + 1))
  fi
}

mkdir -p plugins/nbp-fx plugins/ui digests
touch plugins/nbp-fx/Dockerfile plugins/ui/Dockerfile
printf '{\n  "plugins": []\n}\n' >catalog.json
printf '%s\n' "$da" >digests/nbp-fx

expect_fail "a plugin without a digest" "no signed digest for plugin ui" 2026.11.0 ghcr.io/Example/OpenFief digests catalog.json

printf 'latest\n' >digests/ui
expect_fail "a digest that isn't one" "is not a sha256 digest" 2026.11.0 ghcr.io/Example/OpenFief digests catalog.json

printf '%s\n' "$db" >digests/ui
if ! out="$(run 2026.11.0 ghcr.io/Example/OpenFief digests catalog.json)"; then
  echo "FAIL: adding two plugins: $out" >&2
  failures=$((failures + 1))
fi
want="ghcr.io/example/openfief/nbp-fx@$da 2026.11.0 openfief https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/main
ghcr.io/example/openfief/ui@$db 2026.11.0 openfief https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/main"
got="$(gojq -r '.plugins[] | "\(.image) \(.version) \(.publisher.name) \(.publisher.subject)"' catalog.json)"
if [ "$got" != "$want" ]; then
  echo "FAIL: catalog entries: want $want, got $got" >&2
  failures=$((failures + 1))
fi

expect_fail "the same version again" "already has nbp-fx 2026.11.0" 2026.11.0 ghcr.io/Example/OpenFief digests catalog.json

if ! out="$(run 2026.11.1 ghcr.io/example/openfief digests catalog.json)" || [ "$(gojq '.plugins | length' catalog.json)" != 4 ]; then
  echo "FAIL: a next version keeps the earlier entries: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-catalog: 5 cases passed"
