#!/usr/bin/env bash
# Renders the demo's egress audit (deploy/demo/audit, D-200 criterion 16) for one deploy, after pin-images.sh pinned
# <dir>/openfief: the audit Job runs the egress proxy's image at the digest that verified there, and carries the
# deploy's tag in the annotation openfief.io/deploy, so each deploy's Job differs from the last and Flux re-creates it
# (its force annotation). It fails if the egress proxy's digest in <dir>/openfief is missing or still the placeholder.
# The committed directory stays untouched: the edits go to a copy.
# Usage: render-audit.sh <tag> <out file> [deploy/demo directory]   (kustomize and gojq on PATH)
set -euo pipefail

image='ghcr.io/arkadiuss/openfief/egress-proxy'
placeholder='sha256:0000000000000000000000000000000000000000000000000000000000000000'

fail() {
  echo "render-audit: $*" >&2
  exit 1
}

[ "$#" -ge 2 ] && [ "$#" -le 3 ] || fail "usage: render-audit.sh <tag> <out file> [deploy/demo directory]"
tag="$1"
out="$2"
dir="${3:-deploy/demo}"
[[ "$tag" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,62}$ ]] || fail "'$tag' is not an image tag"
digest="$(gojq -r --yaml-input --arg image "$image" '[.images[]? | select(.name == $image) | .digest] | first // ""' \
  "$dir/openfief/kustomization.yaml")"
[[ "$digest" =~ ^sha256:[0-9a-f]{64}$ ]] || fail "$dir/openfief/kustomization.yaml pins no digest for $image"
[ "$digest" != "$placeholder" ] || fail "$image is still the placeholder in $dir/openfief/kustomization.yaml: run pin-images.sh first"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cp -R "$dir/audit/." "$work/"
(
  cd "$work"
  kustomize edit set image "$image@$digest"
  kustomize edit add annotation "openfief.io/deploy:$tag"
)
kustomize build "$work" >"$out"
echo "render-audit: $out runs $image@$digest for $tag"
