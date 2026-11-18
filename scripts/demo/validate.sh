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

# guard_failures prints one line for each way the demo's spend guards (D-208) are wrong in the built demo with its model layer
# ($1, a manifest stream), and nothing when they are right: the core's flags set the budget (a positive amount), the limits
# per session and per address (10 an hour, 30 a day), 3 calls at once, the proxy whose address header it believes and
# Turnstile's place in its policy; and the built-in UI's config has D-208's limits of one question and a site key, with its
# grant approving the host of turnstileVerifyUrl as an egress host and the Secret ui--turnstile-secret.
guard_failures() {
  local out="$1" args want flag budget
  args="$(gojq -r --yaml-input 'select(.kind == "Deployment" and .metadata.name == "openfief-core") | .spec.template.spec.containers[0].args[]' "$out")"
  for want in --ask-questions-per-hour=10 --ask-questions-per-day=30 --ask-concurrent-calls=3 --allow-turnstile; do
    grep -qxF -- "$want" <<<"$args" || echo "the core's flags lack $want (D-208's guard)"
  done
  for flag in --client-ip-header --trusted-proxies --ask-queue-wait; do
    grep -q -- "^$flag=." <<<"$args" || echo "the core's flags lack $flag (D-208's guards)"
  done
  budget="$(sed -n 's/^--model-daily-budget-usd=//p' <<<"$args")"
  if ! [[ "$budget" =~ ^[0-9]+([.][0-9]+)?$ ]] || ! gojq -n --arg b "$budget" '($b | tonumber) > 0' | grep -q true; then
    echo "the core's --model-daily-budget-usd is not a positive amount of dollars (D-208's daily budget)"
  fi
  gojq -r --yaml-input -s '
    (map(select(.kind == "Plugin" and .metadata.name == "ui"))[0].spec.config // {}) as $c
    | (map(select(.kind == "PluginGrant" and .metadata.name == "ui"))[0].spec // {}) as $g
    | (($c.turnstileVerifyUrl // "") | capture("^https://(?<h>[^/:]+)") | .h) as $host
    | [
        (if ($c.turnstileSiteKey // "") == "" then "the UI has no Turnstile site key (D-208)" else empty end),
        (if ($host // "") == "" or (($g.egress // []) | index($host)) == null
          then "the UI grant does not approve the host of turnstileVerifyUrl as an egress host (D-208, D-75)" else empty end),
        (if (($g.secrets // []) | index("ui--turnstile-secret")) == null then "the UI grant does not name the Secret ui--turnstile-secret (D-208)" else empty end),
        (if $c.askToolRounds != 5 or $c.askToolCalls != 10 or $c.askOutputTokens != 800 or $c.askDeadline != "2m"
          then "the UI'"'"'s limits of one question are not D-208'"'"'s: 5 rounds, 10 tool calls, 800 output tokens and 2m" else empty end)
      ] | .[]' "$out"
}

# grant_failures prints one line for each Secret a Plugin's manifest asks for that the Plugin's PluginGrant does not
# approve, in the live demo's build ($1, a manifest stream), and nothing when every grant covers its manifest. The
# operator names an instance's Secret <instance>--<role> and keeps a Plugin AwaitingApproval, never Ready, while its
# grant leaves out any of them, optional ones included: the demo's openrouter grant once lacked openrouter--ca-bundle
# and only e2e's five-minute wait found it.
grant_failures() {
  gojq -r --yaml-input -s '
    (map(select(.kind == "PluginGrant")) | map({key: .metadata.name, value: (.spec.secrets // [])}) | from_entries) as $grants
    | map(select(.kind == "Plugin"))[]
    | .metadata.name as $n
    | [.spec.manifest.secrets[]? | "\($n)--\(.role)"] - ($grants[$n] // [])
    | .[]
    | "the PluginGrant \($n) does not approve the Secret \(.), which the Plugin \($n)'"'"'s manifest asks for: the Plugin would stay AwaitingApproval"' "$1"
}

# model_layer_failures prints one line for each way the demo's model layer (deploy/demo/model, docs/dev/demo.md) is
# wrong, and nothing when it is right. The layer is a kustomize Component that the live demo (deploy/demo/openfief)
# includes, so the live demo's build is the layer's build and it must: build, with every image pinned by digest; not
# hold the placeholder Turnstile site key; route only the Ask box (the maintainer has no route on the demo); route, and
# allow for the data tiers, only models that its provider instance lists as entries, at that instance; and have D-208's
# spend guards right (guard_failures). Whether the switch is on is not in the manifests: it is the owner's, in Settings.
model_layer_failures() {
  local dir="$1" out="$2" instance="openrouter" listed refs ref
  [ -d "$dir/model" ] || return 0
  if ! kustomize build "$dir/openfief" >"$out" 2>/dev/null; then
    echo "the model layer does not build with the demo ($dir/openfief)"
    return 0
  fi
  if grep -q "REPLACE-WITH-THE-TURNSTILE-SITE-KEY" "$out"; then
    echo "the demo ($dir/openfief) has the placeholder Turnstile site key: the widget is made in Cloudflare and its public key put in $dir/model"
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
  guard_failures "$out"
  grant_failures "$out"
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
