#!/usr/bin/env bash
# Validates the demo's GitOps manifests offline (docs/dev/demo.md): deploy/demo, every git path its Flux
# Kustomizations apply, and the deploy/demo/openfief, deploy/demo/routes and deploy/demo/audit templates CI publishes must build; every image
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

# model_layer_failures prints one line for each way the demo's model layer (deploy/demo/model, docs/dev/demo.md) is
# wrong, and nothing when it is right. The layer is a kustomize Component that the live demo does not list, so it must:
# build with the demo (deploy/demo/with-model) with every image pinned by digest; stay out of the live demo
# (deploy/demo/openfief); route only the Ask box (the maintainer has no route on the demo); and route, and allow for
# the data tiers, only models that its provider instance lists as entries, at that instance.
model_layer_failures() {
  local dir="$1" out="$2" instance="openrouter" listed refs ref
  [ -d "$dir/model" ] || return 0
  if kustomize build "$dir/openfief" 2>/dev/null | grep -q "$instance\|demo-model-settings"; then
    echo "the live demo ($dir/openfief) includes the model layer: it is switched on only on purpose"
  fi
  if ! kustomize build "$dir/with-model" >"$out" 2>/dev/null; then
    echo "the model layer does not build with the demo ($dir/with-model)"
    return 0
  fi
  if [ -n "$(unpinned_images "$out")" ]; then
    echo "the model layer deploys images not pinned by digest: $(unpinned_images "$out" | tr '\n' ' ')"
  fi
  if [ "$(gojq -c '.routes | keys' "$dir/model/settings.json")" != '["ask"]' ]; then
    echo "the model settings route something other than the Ask box"
  fi
  listed="$(gojq -r --yaml-input 'select(.kind == "Plugin" and .metadata.name == "'"$instance"'") | .spec.config.models[].id' "$dir/model/$instance.yaml")"
  refs="$(gojq -r '[.routes[], .tiers[]] | flatten | unique[]' "$dir/model/settings.json")"
  for ref in $refs; do
    if [[ "$ref" != "$instance/"* ]] || ! grep -qxF "${ref#"$instance"/}" <<<"$listed"; then
      echo "the model settings name $ref, which the instance $instance does not list as an entry"
    fi
  done
}

# validate_demo builds every layer and applies the checks; it reports every failure, then fails once.
validate_demo() {
  local dir="${1:-deploy/demo}" path substitutes out failures=0 unverified unknown line
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
    echo "$dir/routes true"
    echo "$dir/audit false"
  )
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    echo "demo: $line" >&2
    failures=$((failures + 1))
  done < <(model_layer_failures "$dir" "$out")
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
