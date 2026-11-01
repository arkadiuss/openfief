#!/usr/bin/env bash
# Tests for scripts/release-bundle.sh (kibi: TEST-tooling-release-bundle), against the real deploy/ and a throwaway
# one. Needs kustomize and gojq on PATH.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
failures=0
# Synthetic digests; not of any real image.
da="sha256:$(printf 'a%.0s' {1..64})"
db="sha256:$(printf 'b%.0s' {1..64})"
dc="sha256:$(printf 'c%.0s' {1..64})"
dd="sha256:$(printf 'd%.0s' {1..64})"
version=2026.11.0-alpha.1

run() { DEPLOY="${deploy:-$root/deploy}" bash "$root/scripts/release-bundle.sh" "$@" 2>&1; }

fail_case() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

expect_fail() {
  local name="$1" pattern="$2" out
  shift 2
  if out="$(run "$@")"; then
    fail_case "$name should fail"
  elif ! grep -qF -- "$pattern" <<<"$out"; then
    fail_case "$name should report '$pattern', got: $out"
  fi
}

# images prints every image the manifests in file $1 run, sorted.
images() {
  gojq -r --yaml-input -s '[.[] | .. | objects | (.containers?, .initContainers?) // empty | .[] | .image] | unique | .[]' "$1"
}

refs=(registry.example.org:5000/e2e/core@"$da" ghcr.io/example/openfief/operator@"$db" ghcr.io/example/openfief/egress-proxy@"$dc")
if ! out="$(run "$version" a "${refs[@]}")"; then
  fail_case "pinning the real install: $out"
else
  want="ghcr.io/example/openfief/egress-proxy@$dc
ghcr.io/example/openfief/operator@$db
registry.example.org:5000/e2e/core@$da"
  got="$(images "a/openfief-install-$version.yaml")"
  [ "$got" = "$want" ] || fail_case "the install's images: want $want, got $got"
  kinds="$(gojq -r --yaml-input -s '[.[] | .kind] | unique | join(" ")' "a/openfief-crds-$version.yaml")"
  [ "$kinds" = "CustomResourceDefinition Namespace" ] || fail_case "the CRDs file holds $kinds"
  [ -z "$(images "a/openfief-crds-$version.yaml")" ] || fail_case "the CRDs file runs an image"
  gojq -e --yaml-input -s 'any(.[]; .kind == "Deployment" and .metadata.name == "openfief-core")' \
    "a/openfief-install-$version.yaml" >/dev/null || fail_case "the install has no core Deployment"
fi

# The same inputs, in another order, give the same bytes.
if ! out="$(run "$version" b "${refs[2]}" "${refs[0]}" "${refs[1]}")"; then
  fail_case "pinning again: $out"
elif ! diff -r a b >/dev/null; then
  fail_case "the same images gave different files"
fi

mkdir digests
printf '%s\n' "$da" >digests/core
printf '%s\n' "$db" >digests/operator
printf '%s\n' "$dc" >digests/egress-proxy
printf '%s\n' "$dd" >digests/nbp-fx
printf '%s\n' "$dd" >digests/demo-seed
if ! out="$(DIGESTS=digests PREFIX=ghcr.io/Example/OpenFief run "$version" c)"; then
  fail_case "pinning from digest files: $out"
else
  want="ghcr.io/example/openfief/core@$da
ghcr.io/example/openfief/egress-proxy@$dc
ghcr.io/example/openfief/operator@$db"
  got="$(images "c/openfief-install-$version.yaml")"
  [ "$got" = "$want" ] || fail_case "the install's images from digest files: want $want, got $got"
fi

# A dry run pins every image to the all-zeros placeholder.
zeros="sha256:$(printf '0%.0s' {1..64})"
if ! out="$(PLACEHOLDER=1 PREFIX=ghcr.io/Example/OpenFief run "$version" e)"; then
  fail_case "pinning placeholders: $out"
else
  want="ghcr.io/example/openfief/core@$zeros
ghcr.io/example/openfief/egress-proxy@$zeros
ghcr.io/example/openfief/operator@$zeros"
  got="$(images "e/openfief-install-$version.yaml")"
  [ "$got" = "$want" ] || fail_case "the install's placeholder images: want $want, got $got"
fi
PLACEHOLDER=1 PREFIX=ghcr.io/example/openfief VERIFY=1 REPO=example/openfief expect_fail \
  "verifying placeholders" "can't be verified" "$version" d
PLACEHOLDER=1 PREFIX=ghcr.io/example/openfief expect_fail "placeholders and images" "takes no images" \
  "$version" d "${refs[@]}"

rm digests/operator
DIGESTS=digests PREFIX=ghcr.io/example/openfief expect_fail "a digest file missing" \
  "no digest for ghcr.io/arkadiuss/openfief/operator" "$version" d
printf 'latest\n' >digests/operator
DIGESTS=digests PREFIX=ghcr.io/example/openfief expect_fail "a digest file that isn't one" \
  "is not a sha256 digest" "$version" d
DIGESTS=digests expect_fail "digest files without a prefix" "DIGESTS needs PREFIX" "$version" d
DIGESTS=digests PREFIX=ghcr.io/example/openfief expect_fail "digest files and arguments" "not both" \
  "$version" d "${refs[@]}"

expect_fail "an image missing" "no digest for ghcr.io/arkadiuss/openfief/egress-proxy" "$version" d "${refs[0]}" "${refs[1]}"
expect_fail "an image the install doesn't name" "the install names no demo-seed image" \
  "$version" d "${refs[@]}" ghcr.io/example/openfief/demo-seed@"$dd"
expect_fail "an image by tag" "is not <repository>@sha256:<digest>" "$version" d ghcr.io/example/openfief/core:2026.11.0
expect_fail "two images for one name" "two images for core" "$version" d "${refs[@]}" ghcr.io/example/x/core@"$dd"
expect_fail "a version that can't name a file" "is not a version" "../x" d "${refs[@]}"
VERIFY=1 expect_fail "verifying without the repository" "VERIFY=1 needs REPO" "$version" d "${refs[@]}"
[ ! -e d ] || fail_case "a failed run wrote files"

# A base that names its image by tag can't be pinned without hiding the tag.
mkdir -p fake/crds fake/operator fake/egress-proxy fake/core
cp "$root/deploy/crds/"* fake/crds/
for base in operator egress-proxy core; do
  printf 'resources:\n  - deployment.yaml\n' >"fake/$base/kustomization.yaml"
  printf 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\nspec:\n  template:\n    spec:\n      containers:\n        - name: %s\n          image: ghcr.io/arkadiuss/openfief/%s\n' \
    "$base" "$base" "$base" >"fake/$base/deployment.yaml"
done
sed -i 's|openfief/core$|openfief/core:latest|' fake/core/deployment.yaml
deploy="$work/fake" expect_fail "a base image with a tag" "with a tag or digest" "$version" d "${refs[@]}"

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-bundle: 18 cases passed"
