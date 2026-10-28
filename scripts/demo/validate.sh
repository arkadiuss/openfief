#!/usr/bin/env bash
# Validates the demo's GitOps manifests offline (docs/dev/demo.md): deploy/demo, every git path its Flux
# Kustomizations apply, and the deploy/demo/openfief template CI publishes must build with kustomize; every image
# they deploy must be pinned by digest (D-17); and every OCI source must require a cosign signature.
# Usage: validate.sh [deploy/demo directory]   (kustomize and gojq on PATH)
set -euo pipefail

# stack_paths prints the repository paths that the git-sourced Flux Kustomizations in a stack file apply.
stack_paths() {
  gojq -r --yaml-input 'select(.kind == "Kustomization" and .spec.sourceRef.kind == "GitRepository") | .spec.path' "$1"
}

# unpinned_images prints each image in a manifest stream that is not pinned by a sha256 digest.
unpinned_images() {
  gojq -r --yaml-input '.. | objects | select(has("image")) | .image | strings' "$1" |
    grep -Ev '@sha256:[0-9a-f]{64}$' || true
}

# validate_demo builds every layer and applies the checks; it reports every failure, then fails once.
validate_demo() {
  local dir="${1:-deploy/demo}" path out failures=0 unverified
  out="$(mktemp)"
  if ! kustomize build "$dir" >"$out"; then
    echo "demo: kustomize build $dir failed" >&2
    failures=$((failures + 1))
  fi
  for path in $(stack_paths "$dir/stack.yaml") "$dir/openfief"; do
    if ! kustomize build "$path" >"$out"; then
      echo "demo: kustomize build $path failed" >&2
      failures=$((failures + 1))
      continue
    fi
    if [ -n "$(unpinned_images "$out")" ]; then
      echo "demo: $path deploys images not pinned by digest: $(unpinned_images "$out" | tr '\n' ' ')" >&2
      failures=$((failures + 1))
    fi
  done
  rm -f "$out"
  unverified="$(gojq -r --yaml-input 'select(.kind == "OCIRepository" and .spec.verify.provider != "cosign") | .metadata.name' "$dir/stack.yaml")"
  if [ -n "$unverified" ]; then
    echo "demo: OCI sources without cosign verification: $unverified" >&2
    failures=$((failures + 1))
  fi
  if [ "$failures" -ne 0 ]; then
    return 1
  fi
  echo "demo: manifests build and every image is pinned by digest"
}

validate_demo "$@"
