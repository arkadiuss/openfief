#!/usr/bin/env bash
# Tests for scripts/release-bundle.sh (kibi: TEST-tooling-release-bundle), against the real deploy/ and a throwaway
# one, with a synthetic CloudNativePG manifest. Needs kustomize and gojq on PATH.
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
# A synthetic stand-in for CloudNativePG's release manifest, with its sum as the pin.
printf 'apiVersion: v1\nkind: Namespace\nmetadata:\n  name: cnpg-system\n' >cnpg-1.2.3.yaml
cnpg_sum="$(sha256sum cnpg-1.2.3.yaml | cut -d' ' -f1)"

run() {
  DEPLOY="${deploy:-$root/deploy}" CNPG="${CNPG-$work/cnpg-1.2.3.yaml}" CNPG_SHA256="${CNPG_SHA256-$cnpg_sum}" \
    bash "$root/scripts/release-bundle.sh" "$@" 2>&1
}

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
  [ "$kinds" = "CustomResourceDefinition LimitRange Namespace" ] || fail_case "the CRDs file holds $kinds"
  [ -z "$(images "a/openfief-crds-$version.yaml")" ] || fail_case "the CRDs file runs an image"
  gojq -e --yaml-input -s 'any(.[]; .kind == "Deployment" and .metadata.name == "openfief-core")' \
    "a/openfief-install-$version.yaml" >/dev/null || fail_case "the install has no core Deployment"
  # The database the install brings by default (D-14): exactly the cluster the install reads, and nothing it runs.
  database="$(gojq -r --yaml-input -s '[.[] | select(. != null) | "\(.apiVersion) \(.kind) \(.metadata.namespace)/\(.metadata.name) \(.spec.bootstrap.initdb.database)/\(.spec.bootstrap.initdb.owner)"] | join(",")' \
    "a/openfief-db-$version.yaml")"
  [ "$database" = "postgresql.cnpg.io/v1 Cluster openfief/openfief-db openfief/openfief" ] || fail_case "the database file holds $database"
  # Its own requests and memory limit (D-182), so the namespace's LimitRange defaults don't size it.
  sized="$(gojq -r --yaml-input -s '[.[] | select(. != null) | .spec.resources | "\(.requests.cpu) \(.requests.memory) \(.limits.memory)"] | join(",")' \
    "a/openfief-db-$version.yaml")"
  [ "$sized" = "100m 256Mi 1Gi" ] || fail_case "the database's requests and memory limit are $sized, want 100m 256Mi 1Gi"
  grep -q "^#   kubectl apply --server-side -f openfief-db-$version.yaml" "a/openfief-db-$version.yaml" ||
    fail_case "the database file doesn't say how to apply it"
  cmp -s cnpg-1.2.3.yaml a/cnpg-1.2.3.yaml || fail_case "the CloudNativePG manifest isn't shipped byte for byte"
  [ "$(find a -type f | wc -l)" -eq 4 ] || fail_case "the bundle holds $(find a -type f | wc -l) files, want 4"
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
CNPG= expect_fail "no CloudNativePG manifest" "CNPG and CNPG_SHA256 must" "$version" d "${refs[@]}"
CNPG_SHA256="$(printf 'f%.0s' {1..64})" expect_fail "a CloudNativePG manifest that isn't the pinned one" \
  "not the pinned" "$version" d "${refs[@]}"
CNPG_SHA256=latest expect_fail "a pin that isn't a sha256" "is not a sha256" "$version" d "${refs[@]}"
cp cnpg-1.2.3.yaml operator.yaml
CNPG="$work/operator.yaml" expect_fail "a CloudNativePG manifest under another name" "is not named cnpg-" \
  "$version" d "${refs[@]}"
CNPG="$work/cnpg-9.9.9.yaml" expect_fail "a CloudNativePG manifest that doesn't exist" "is not a file" \
  "$version" d "${refs[@]}"
[ ! -e d ] || fail_case "a failed run wrote files"

# A base that names its image by tag can't be pinned without hiding the tag.
mkdir -p fake/crds fake/operator fake/egress-proxy fake/core fake/database
cp "$root/deploy/crds/"* fake/crds/
cp "$root/deploy/database/"* fake/database/
for base in operator egress-proxy core; do
  printf 'resources:\n  - deployment.yaml\n' >"fake/$base/kustomization.yaml"
  printf 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\nspec:\n  template:\n    spec:\n      containers:\n        - name: %s\n          image: ghcr.io/arkadiuss/openfief/%s\n' \
    "$base" "$base" "$base" >"fake/$base/deployment.yaml"
done
sed -i 's|openfief/core$|openfief/core:latest|' fake/core/deployment.yaml
deploy="$work/fake" expect_fail "a base image with a tag" "with a tag or digest" "$version" d "${refs[@]}"
sed -i 's|openfief/core:latest$|openfief/core|' fake/core/deployment.yaml
# The real install sets no exemption from the proxy's refused ranges (D-215); one that does, as a flag or as an
# environment variable, fails the bundle.
if grep -qiE 'refused-range-exempt-host|REFUSED_RANGE_EXEMPT_HOST' "a/openfief-install-$version.yaml"; then
  fail_case "the real install sets the egress proxy's test-only exemption"
fi
printf '          args: ["-listen=:4750", "-refused-range-exempt-host=fixture.example.test=10.0.0.1"]\n' >>fake/egress-proxy/deployment.yaml
deploy="$work/fake" expect_fail "an exemption set by a flag" "test-only exemption from its refused ranges (D-215)" \
  "$version" d "${refs[@]}"
[ ! -e d ] || fail_case "a bundle with an exemption wrote files"
sed -i '$d' fake/egress-proxy/deployment.yaml
printf '          env:\n            - name: OPENFIEF_REFUSED_RANGE_EXEMPT_HOST\n              value: fixture.example.test=10.0.0.1\n' >>fake/egress-proxy/deployment.yaml
deploy="$work/fake" expect_fail "an exemption set by an environment variable" "OPENFIEF_REFUSED_RANGE_EXEMPT_HOST" \
  "$version" d "${refs[@]}"
[ ! -e d ] || fail_case "a bundle with an exemption wrote files"
sed -i '$d;$d;$d' fake/egress-proxy/deployment.yaml
sed -i 's|name: openfief-db$|name: other-db|' fake/database/openfief-db.yaml
deploy="$work/fake" expect_fail "a database the install doesn't read" "exactly the Cluster openfief/openfief-db" \
  "$version" d "${refs[@]}"
[ ! -e d ] || fail_case "a refused database wrote files"

# The install guide names the CloudNativePG release and the sha256 the Taskfile pins, which the bundle ships.
pinned_version="$(sed -n 's/^  CNPG_VERSION: *//p' "$root/Taskfile.yml")"
pinned_sum="$(sed -n 's/^  CNPG_MANIFEST_SHA256: *//p' "$root/Taskfile.yml")"
[ -n "$pinned_version" ] && [ -n "$pinned_sum" ] || fail_case "the Taskfile pins no CloudNativePG manifest"
grep -qF "cnpg-$pinned_version.yaml" "$root/docs/guides/install.md" ||
  fail_case "the install guide doesn't name cnpg-$pinned_version.yaml"
grep -qF "$pinned_sum" "$root/docs/guides/install.md" || fail_case "the install guide doesn't give the pinned sha256"

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-bundle: 29 cases passed"
