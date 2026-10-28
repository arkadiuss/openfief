#!/usr/bin/env bash
# Pins the demo's images to signed digests (docs/dev/demo.md#deploys, D-17). For each image listed under
# `images:` in <dir>/openfief/kustomization.yaml it verifies <image>:<tag> with cosign against the keyless
# identity of this repository's images workflow on main, and writes the digest that verified. It fails if any
# image doesn't verify or a placeholder digest is left. CI runs it on its own checkout, then publishes the result.
# Usage: pin-images.sh <tag> <owner/repo> [deploy/demo directory]
set -euo pipefail

placeholder='sha256:0000000000000000000000000000000000000000000000000000000000000000'

# image_names prints the image names listed under `images:` in a kustomization.
image_names() {
  gojq -r --yaml-input '.images[]?.name' "$1"
}

# verified_digest prints the digest of <image>:<tag> whose cosign signature comes from the repository's images
# workflow on main; it fails when nothing verifies.
verified_digest() {
  local image="$1" tag="$2" repo="$3" digest
  digest="$(cosign verify "$image:$tag" \
    --certificate-identity "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main" \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    --output json 2>/dev/null | gojq -r '[.[].critical.image["docker-manifest-digest"]] | unique | .[]')" || true
  if [[ ! "$digest" =~ ^sha256:[0-9a-f]{64}$ ]]; then
    echo "pin-images: no verified signature for $image:$tag from $repo's images workflow" >&2
    return 1
  fi
  echo "$digest"
}

# set_digest rewrites the digest line that follows `- name: <image>` in a kustomization's images list.
set_digest() {
  local file="$1" image="$2" digest="$3" out
  out="$(awk -v image="$image" -v digest="$digest" '
    $0 ~ "^  - name: " && $3 == image { found = 1; print; next }
    found && $1 == "digest:" { sub(/digest: .*/, "digest: " digest); found = 0; done = 1 }
    { print }
    END { if (!done) exit 1 }' "$file")" || {
    echo "pin-images: no digest line for $image in $file" >&2
    return 1
  }
  printf '%s\n' "$out" >"$file"
}

# pin_images pins every listed image and fails if a placeholder is left.
pin_images() {
  local tag="$1" repo="$2" dir="${3:-deploy/demo}" kustomization image digest
  if [[ ! "$tag" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$ ]] || [[ ! "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
    echo "pin-images: usage: pin-images.sh <tag> <owner/repo> [dir]" >&2
    return 1
  fi
  kustomization="$dir/openfief/kustomization.yaml"
  for image in $(image_names "$kustomization"); do
    digest="$(verified_digest "$image" "$tag" "$repo")"
    set_digest "$kustomization" "$image" "$digest"
    echo "pin-images: $image:$tag -> $digest"
  done
  if grep -q "$placeholder" "$kustomization"; then
    echo "pin-images: a placeholder digest is left in $kustomization" >&2
    return 1
  fi
}

pin_images "$@"
