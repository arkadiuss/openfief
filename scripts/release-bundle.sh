#!/usr/bin/env bash
# Renders a release's bundle (docs/dev/sdlc.md#releases): the CRDs and the bootstrap install, with every image the
# install names pinned to the digest the release signed (D-17), and the database the install brings by default (D-14),
# as plain manifests:
#   <out dir>/openfief-crds-<version>.yaml     deploy/crds: the openfief namespace with its LimitRange (D-182) and
#                                              the Plugin and PluginGrant CRDs, the part that needs a cluster admin
#                                              (D-15)
#   <out dir>/openfief-install-<version>.yaml  deploy/operator, deploy/egress-proxy, deploy/core and deploy/spicedb
#   <out dir>/openfief-db-<version>.yaml       deploy/database: a minimal CloudNativePG cluster openfief-db and the
#                                              database SpiceDB uses in it (D-164), for the owner to edit
#   <out dir>/cnpg-<cnpg version>.yaml         CloudNativePG's own release manifest, byte for byte: CNPG=<file> names
#                                              it, CNPG_SHA256 is its pinned sha256 (the Taskfile's
#                                              CNPG_MANIFEST_SHA256), and a file with any other sum is refused. An
#                                              owner whose cluster already runs CloudNativePG skips it.
# Plugins aren't in it: they are installed with the CLI (D-46), and the catalog the operator and the CLI embed pins
# each first-party plugin's signed digest (D-88).
#
# The images come in one of two ways:
#   - as arguments, each <repository>@sha256:<digest>, matched to the install's image of the same last path element
#     (the core's ghcr.io/arkadiuss/openfief/core takes <anything>/core@sha256:...); an argument the install doesn't
#     name is refused, so a typo can't leave an image unpinned. task e2e passes the suite's registry this way.
#   - with DIGESTS=<dir> and PREFIX=<registry prefix>, from the digest files images.yml writes, one per image, named
#     after it: <prefix>/<name>@<digest>. Files for images the install doesn't name (plugins, demo-seed) are ignored.
#   - with PLACEHOLDER=1 and PREFIX, every image at an all-zeros digest: a release's dry run, which pushes and signs
#     nothing, renders the bundle this way so that everything but the digests is checked.
# An image the install names that already carries its digest is a third party's (SpiceDB, the Job that generates its
# credentials): it stays as the base pins it, and is neither given nor verified here, since only images.yml signs ours.
# Every other image the install names must get a digest. With VERIFY=1 and REPO=<owner/repo> (the release workflow), each
# pinned image must also carry a keyless cosign signature from that repository's images.yml on main, for that same
# digest (scripts/verified-digest.sh, which needs cosign and gojq).
# The output depends only on deploy/, the images and the CloudNativePG manifest given, so the same inputs give
# byte-identical files.
# Needs kustomize and gojq on PATH. DEPLOY overrides the deploy/ directory (tests).
# Usage: release-bundle.sh <version> <out dir> [image...]
set -euo pipefail
shopt -s inherit_errexit

usage='usage: CNPG=<cnpg-X.Y.Z.yaml> CNPG_SHA256=<sha256> release-bundle.sh <version> <out dir> [image...]'
if [ "$#" -lt 2 ]; then
  echo "$usage" >&2
  exit 2
fi
version="$1"
out="$2"
shift 2
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
deploy="${DEPLOY:-$root/deploy}"
digest_re='sha256:[0-9a-f]{64}'
# The digest PLACEHOLDER=1 pins every image to, as the demo's template does (deploy/demo/openfief).
placeholder="sha256:$(printf '0%.0s' {1..64})"

fail() {
  echo "release-bundle: $*" >&2
  exit 1
}

# The version names the files, so it may hold only what a file name and a tag can.
[[ "$version" =~ ^[0-9A-Za-z][0-9A-Za-z.-]*$ ]] || fail "'$version' is not a version"

# check_cnpg fails unless CNPG names CloudNativePG's release manifest, cnpg-<version>.yaml as upstream names it,
# with the pinned sha256 CNPG_SHA256.
check_cnpg() {
  [ -n "${CNPG:-}" ] && [ -n "${CNPG_SHA256:-}" ] ||
    fail "CNPG and CNPG_SHA256 must name CloudNativePG's release manifest and its pinned sha256 (task release:bundle)"
  [[ "$(basename "$CNPG")" =~ ^cnpg-[0-9]+\.[0-9]+\.[0-9]+\.yaml$ ]] ||
    fail "CNPG: $(basename "$CNPG") is not named cnpg-<version>.yaml, as CloudNativePG's release names it"
  [[ "$CNPG_SHA256" =~ ^[0-9a-f]{64}$ ]] || fail "CNPG_SHA256: '$CNPG_SHA256' is not a sha256"
  [ -f "$CNPG" ] || fail "CNPG: $CNPG is not a file"
  local sum
  sum="$(sha256sum "$CNPG" | cut -d' ' -f1)"
  [ "$sum" = "$CNPG_SHA256" ] || fail "CNPG: $CNPG has sha256 $sum, not the pinned $CNPG_SHA256"
}

# check_database fails unless the manifests in file $1 are exactly one CloudNativePG Cluster openfief-db in namespace
# openfief, the cluster the install reads its database from, and the Database spicedb in it (D-164), and run no image
# of their own.
check_database() {
  local objects
  objects="$(gojq -r --yaml-input -s '[.[] | select(. != null) | "\(.kind)/\(.metadata.namespace)/\(.metadata.name)"] | join(" ")' "$1")"
  [ "$objects" = "Cluster/openfief/openfief-db Database/openfief/openfief-db-spicedb" ] ||
    fail "deploy/database must hold exactly the Cluster openfief/openfief-db and the Database openfief/openfief-db-spicedb, not: $objects"
  [ -z "$(install_images "$1")" ] || fail "deploy/database runs an image of its own"
}

# install_images prints each image the manifests in file $1 run, once, sorted.
install_images() {
  gojq -r --yaml-input -s \
    '[.[] | .. | objects | (.containers?, .initContainers?) // empty | .[] | .image] | unique | .[]' "$1"
}

# check_no_exemption fails if a manifest in file $1 sets the egress proxy's test-only exemption from its refused
# address ranges (D-215), as a flag or as an environment variable, so a shipped install never has one.
check_no_exemption() {
  local found
  found="$(gojq -r --yaml-input -s \
    '[.[] | .. | objects | (.containers?, .initContainers?) // empty | .[]
      | ((.args? // [])[], (.command? // [])[], ((.env? // [])[] | .name)) | select(type == "string")
      | select(test("refused-range-exempt-host|REFUSED_RANGE_EXEMPT_HOST"; "i"))] | unique | join(", ")' "$1")"
  [ -z "$found" ] || fail "the install sets the egress proxy's test-only exemption from its refused ranges (D-215): $found"
}

# third_party is the images the install names with their digest already (name -> 1): a third party's, which the bundle
# leaves as the base pins them.
declare -A third_party=()

# given_images fills the array `given` (name -> image) from the arguments, or from DIGESTS and PREFIX.
declare -A given=()
given_images() {
  local image name digest prefix file
  if [ "${PLACEHOLDER:-0}" = 1 ]; then
    [ "$#" -eq 0 ] && [ -z "${DIGESTS:-}" ] || fail "PLACEHOLDER=1 takes no images"
    [ "${VERIFY:-0}" != 1 ] || fail "PLACEHOLDER=1 pins nothing that was signed, so it can't be verified"
    [ -n "${PREFIX:-}" ] || fail "PLACEHOLDER=1 needs PREFIX, the registry prefix to name the images under"
    return
  fi
  if [ -n "${DIGESTS:-}" ]; then
    [ "$#" -eq 0 ] || fail "give the images as arguments or with DIGESTS, not both"
    [ -n "${PREFIX:-}" ] || fail "DIGESTS needs PREFIX, the registry prefix the images were pushed under"
    [ -d "$DIGESTS" ] || fail "DIGESTS: $DIGESTS is not a directory"
    prefix="$(tr '[:upper:]' '[:lower:]' <<<"$PREFIX")"
    for file in "$DIGESTS"/*; do
      [ -f "$file" ] || continue
      name="$(basename "$file")"
      digest="$(tr -d '[:space:]' <"$file")"
      [[ "$digest" =~ ^${digest_re}$ ]] || fail "$file: '$digest' is not a sha256 digest"
      given[$name]="$prefix/$name@$digest"
    done
    return
  fi
  for image in "$@"; do
    [[ "$image" =~ ^[a-z0-9][a-z0-9._:/-]*/([a-z0-9][a-z0-9._-]*)@${digest_re}$ ]] ||
      fail "'$image' is not <repository>@sha256:<digest>"
    name="${BASH_REMATCH[1]}"
    [ -z "${given[$name]:-}" ] || fail "two images for $name: ${given[$name]} and $image"
    given[$name]="$image"
  done
}

# pin writes to $2 a kustomization that sets each image of base file $1 to its given digest, and prints nothing.
# It fails when the install names an image that isn't a bare repository, or one without a given digest.
pin() {
  local base="$1" kustomization="$2" image name ref
  declare -A used=()
  printf 'apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n  - %s\nimages:\n' \
    "$(basename "$base")" >"$kustomization"
  while IFS= read -r image; do
    name="${image##*/}"
    if [[ "$image" =~ ^[a-z0-9][a-z0-9._:/-]*:[A-Za-z0-9_][A-Za-z0-9_.-]*@${digest_re}$ ]]; then
      third_party[$image]=1
      continue
    fi
    [[ "$image" != *@* && "$name" != *:* ]] ||
      fail "deploy/ names $image with a tag or digest; its bases name bare repositories, which the bundle pins (a third party's image names a tag and its digest, both)"
    if [ "${PLACEHOLDER:-0}" = 1 ]; then
      given[$name]="$(tr '[:upper:]' '[:lower:]' <<<"$PREFIX")/$name@$placeholder"
    fi
    ref="${given[$name]:-}"
    [ -n "$ref" ] || fail "no digest for $image"
    used[$name]=1
    printf '  - name: %s\n    newName: %s\n    digest: %s\n' "$image" "${ref%@*}" "${ref#*@}" >>"$kustomization"
  done < <(install_images "$base")
  [ "${#used[@]}" -gt 0 ] || fail "deploy/ names no image of ours"
  if [ -z "${DIGESTS:-}" ]; then
    for name in "${!given[@]}"; do
      [ -n "${used[$name]:-}" ] || fail "the install names no $name image: ${given[$name]}"
    done
  fi
}

# check_pinned fails unless every image the manifests in file $1 run is one of the given images, by digest.
check_pinned() {
  local image name
  while IFS= read -r image; do
    [ -z "${third_party[$image]:-}" ] || continue
    name="${image##*/}"
    name="${name%@*}"
    [ "$image" = "${given[$name]:-}" ] || fail "the install runs $image, not a given digest"
  done < <(install_images "$1")
}

# verify fails unless each image the manifests in file $1 run verifies as images.yml on main of REPO signed it.
verify() {
  local image digest
  [ -n "${REPO:-}" ] || fail "VERIFY=1 needs REPO, the owner/repo whose images workflow signed the images"
  while IFS= read -r image; do
    [ -z "${third_party[$image]:-}" ] || continue
    digest="$(bash "$root/scripts/verified-digest.sh" "$image" "$REPO")" ||
      fail "$image has no signature from $REPO's images workflow on main"
    [ "$digest" = "${image#*@}" ] || fail "$image: the signature names $digest"
  done < <(install_images "$1")
}

check_cnpg
given_images "$@"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir "$work/install"
kustomize build "$deploy/crds" >"$work/crds.yaml"
kustomize build "$deploy/database" >"$work/database.yaml"
check_database "$work/database.yaml"
for base in operator egress-proxy core spicedb; do
  printf -- '---\n'
  kustomize build "$deploy/$base"
done >"$work/install/base.yaml"
pin "$work/install/base.yaml" "$work/install/kustomization.yaml"
kustomize build "$work/install" >"$work/install.yaml"
check_pinned "$work/install.yaml"
check_no_exemption "$work/install.yaml"
if [ "${VERIFY:-0}" = 1 ]; then
  verify "$work/install.yaml"
fi

mkdir -p "$out"
{
  printf '# OpenFief %s: the namespace and the CRDs. Apply first, as a cluster admin:\n' "$version"
  printf '#   kubectl apply --server-side -f openfief-crds-%s.yaml\n' "$version"
  cat "$work/crds.yaml"
} >"$out/openfief-crds-$version.yaml"
{
  printf "# OpenFief %s: the operator, the egress proxy, the core and SpiceDB. Our images are pinned to their signed digests, and\n" "$version"
  printf "# SpiceDB's and its credentials Job's third-party images to the digests this release was built with.\n"
  printf '# Apply after openfief-crds-%s.yaml, next to a CloudNativePG cluster openfief-db in namespace openfief:\n' "$version"
  printf '#   kubectl apply --server-side -f openfief-install-%s.yaml\n' "$version"
  cat "$work/install.yaml"
} >"$out/openfief-install-$version.yaml"
{
  printf '# OpenFief %s: its database, the smallest CloudNativePG cluster that works (D-14): one instance on the\n' "$version"
  printf "# cluster's default storage class. Edit it to suit your cluster (size, storageClass, CloudNativePG's backups), but\n"
  printf '# keep its name, namespace, database and owner: the install reads the Secret openfief-db-app CloudNativePG makes.\n'
  printf "# The managed role spicedb and the Database next to the Cluster are SpiceDB's (D-164): keep them too.\n"
  printf '# Apply after %s and openfief-crds-%s.yaml:\n' "$(basename "$CNPG")" "$version"
  printf '#   kubectl apply --server-side -f openfief-db-%s.yaml\n' "$version"
  cat "$work/database.yaml"
} >"$out/openfief-db-$version.yaml"
cp "$CNPG" "$out/$(basename "$CNPG")"
echo "release-bundle: wrote $out/openfief-crds-$version.yaml, $out/openfief-install-$version.yaml," \
  "$out/openfief-db-$version.yaml and $out/$(basename "$CNPG")" >&2
install_images "$out/openfief-install-$version.yaml" | sed 's/^/release-bundle: pinned /' >&2
