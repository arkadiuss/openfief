#!/usr/bin/env bash
# Validates the demo's GitOps manifests offline (docs/dev/demo.md): deploy/demo, every git path its Flux
# Kustomizations apply, and the deploy/demo/openfief template CI publishes must build with kustomize; every image
# they deploy must be pinned by digest (D-17); every OCI source must require a cosign signature; and a layer with
# Flux variable substitution may only use the variables the machine's cloud-init provides.
# Usage: validate.sh [deploy/demo directory]   (kustomize and gojq on PATH)
set -euo pipefail

# The variables cloud-init writes into the ConfigMap flux-system/demo-settings (deploy/demo/cloud-init.yaml).
settings='DEMO_HOST DEMO_ACME_SERVER'

# stack_paths prints the repository path that each git-sourced Flux Kustomization in a stack file applies, and
# whether it substitutes variables (true or false).
stack_paths() {
  gojq -r --yaml-input 'select(.kind == "Kustomization" and .spec.sourceRef.kind == "GitRepository")
    | "\(.spec.path) \(.spec.postBuild != null)"' "$1"
}

# unpinned_images prints each image in a manifest stream that is not pinned by a sha256 digest.
unpinned_images() {
  gojq -r --yaml-input '.. | objects | select(has("image")) | .image | strings' "$1" |
    grep -Ev '@sha256:[0-9a-f]{64}$' || true
}

# unknown_variables prints each ${NAME} in a manifest stream that the demo settings don't define: Flux would
# replace it with an empty string.
unknown_variables() {
  local name
  for name in $(grep -oE '[$][{][A-Za-z_][A-Za-z0-9_]*[}]' "$1" | sort -u | tr -d '${}'); do
    [[ " $settings " == *" $name "* ]] || echo "$name"
  done
}

# validate_demo builds every layer and applies the checks; it reports every failure, then fails once.
validate_demo() {
  local dir="${1:-deploy/demo}" path substitutes out failures=0 unverified unknown
  out="$(mktemp)"
  if ! kustomize build "$dir" >"$out"; then
    echo "demo: kustomize build $dir failed" >&2
    failures=$((failures + 1))
  fi
  while read -r path substitutes; do
    if ! kustomize build "$path" >"$out"; then
      echo "demo: kustomize build $path failed" >&2
      failures=$((failures + 1))
      continue
    fi
    if [ -n "$(unpinned_images "$out")" ]; then
      echo "demo: $path deploys images not pinned by digest: $(unpinned_images "$out" | tr '\n' ' ')" >&2
      failures=$((failures + 1))
    fi
    unknown="$(unknown_variables "$out" | tr '\n' ' ')"
    if [ "$substitutes" = true ] && [ -n "$unknown" ]; then
      echo "demo: $path uses variables the demo settings don't define: $unknown" >&2
      failures=$((failures + 1))
    fi
  done < <(
    stack_paths "$dir/stack.yaml"
    echo "$dir/openfief false"
  )
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
