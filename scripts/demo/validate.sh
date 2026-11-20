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

# telemetry_failures prints one line for each way the demo's telemetry layer (deploy/demo/telemetry, docs/dev/demo.md#telemetry)
# is wrong, and nothing when it is right. The layer ships the demo's logs, traces and metrics to Grafana Cloud, so visitors'
# personal data must not leave (GR-12), and it must send nowhere else. $1 is the deploy/demo directory, $2 the build of
# the layer and $3 the build of the live demo (deploy/demo/openfief). It checks that:
#   - Flux applies the layer, and the build holds no Secret (the credentials are the Secret grafana-otlp, made out of band);
#   - every pipeline of the collector's configuration, and there is one for logs, traces and metrics, runs the processor
#     transform/scrub (with nothing but the memory limiter before it) and exports to the one exporter otlphttp/grafana;
#   - transform/scrub fails closed (error_mode propagate) and drops keys and rewrites values in every context a signal's
#     attributes live in, the log body and the span's name and status message included, with the keys and patterns of
#     the personal data it exists for;
#   - the exporter's endpoint and the basic auth's instance ID and token come from the environment, so no account's address
#     or credential is in git, and the ConfigMap the DaemonSet mounts is the file;
#   - the DaemonSet's mounts are read-only but for its own directory of positions, it holds no capability, and it takes
#     the credentials from the Secret's three keys without optional, so it waits for the Secret instead of running without;
#   - the NetworkPolicy lets only the core and the operator reach the OTLP port, and lets the collector reach only DNS and
#     HTTPS to the public internet, with the private ranges left out;
#   - the live demo points the core and the operator at the collector's Service, and the core's policy at that port alone.
telemetry_failures() {
  local dir="$1" built="$2" live="$3" config="$1/telemetry/collector.yaml"
  if ! gojq -e --yaml-input 'select(.kind == "Kustomization" and .spec.sourceRef.kind == "GitRepository" and .spec.path == "./deploy/demo/telemetry")' "$dir/stack.yaml" >/dev/null 2>&1; then
    echo "the demo's Flux stack does not apply ./deploy/demo/telemetry (docs/dev/demo.md#telemetry)"
  fi
  if grep -q 'grafana\.net' "$config"; then
    echo "the collector's configuration names a Grafana address: the endpoint comes from the Secret grafana-otlp, so no account's address is in git"
  fi
  gojq -r --yaml-input '
    def stmts($ctx; $prefix): [.[] | select(.context == $ctx) | (.statements // [])[] | select(startswith($prefix))];
    . as $c
    | ($c.processors["transform/scrub"] // {}) as $s
    | (
        ($c.service.pipelines // {} | to_entries[]
          | .key as $k | .value as $p
          | (if (($p.processors // []) | index("transform/scrub")) == null
              then "the pipeline \($k) does not run the processor transform/scrub: personal data would reach Grafana (GR-12)" else empty end),
            (if (($p.processors // []) | index("transform/scrub")) as $i | $i != null and ($p.processors[:$i] | map(select(. != "memory_limiter")) | length) > 0
              then "the pipeline \($k) runs a processor other than the memory limiter before transform/scrub" else empty end),
            (if ($p.exporters // []) != ["otlphttp/grafana"]
              then "the pipeline \($k) exports somewhere other than the one exporter otlphttp/grafana" else empty end)),
        (["logs", "traces", "metrics"][] as $sig
          | if ([$c.service.pipelines // {} | keys[] | select(split("/")[0] == $sig)] | length) == 0
              then "the collector has no \($sig) pipeline" else empty end),
        (if ($c.exporters // {} | keys) != ["otlphttp/grafana"] then "the collector has an exporter other than otlphttp/grafana" else empty end),
        (if ($c.exporters["otlphttp/grafana"].endpoint // "") != "${env:GRAFANA_OTLP_ENDPOINT}"
          then "the exporter otlphttp/grafana does not take its endpoint from ${env:GRAFANA_OTLP_ENDPOINT}" else empty end),
        (if ($c.exporters["otlphttp/grafana"].auth.authenticator // "") != "basicauth/grafana"
          then "the exporter otlphttp/grafana does not authenticate with basicauth/grafana" else empty end),
        (if ($c.extensions["basicauth/grafana"].client_auth.username // "") != "${env:GRAFANA_INSTANCE_ID}"
            or ($c.extensions["basicauth/grafana"].client_auth.password // "") != "${env:GRAFANA_OTLP_TOKEN}"
          then "basicauth/grafana does not take its username and password from ${env:GRAFANA_INSTANCE_ID} and ${env:GRAFANA_OTLP_TOKEN}" else empty end),
        (if ($s.error_mode // "") != "propagate" then "transform/scrub is not fail-closed: its error_mode is not propagate" else empty end),
        ({log_statements: ["resource", "scope", "log"], trace_statements: ["resource", "scope", "span", "spanevent"], metric_statements: ["resource", "scope", "datapoint"]}
          | to_entries[] | .key as $kind | .value[] as $ctx
          | ($s[$kind] // []) as $blocks
          | (if ($blocks | stmts($ctx; "delete_matching_keys(\($ctx).attributes, ") | length) == 0
              then "transform/scrub has no delete_matching_keys of \($ctx).attributes in its \($kind)" else empty end),
            (if ($blocks | stmts($ctx; "replace_all_patterns(\($ctx).attributes, \"value\", ") | length) == 0
              then "transform/scrub has no replace_all_patterns of \($ctx).attributes values in its \($kind)" else empty end)),
        (if (($s.log_statements // []) | stmts("log"; "replace_pattern(log.body, ") | length) == 0
          then "transform/scrub does not rewrite the log body" else empty end),
        (["span.name", "span.status.message"][] as $f
          | if (($s.trace_statements // []) | stmts("span"; "replace_pattern(\($f), ") | length) == 0
              then "transform/scrub does not rewrite \($f)" else empty end),
        ([($s | .. | strings | select(startswith("delete_matching_keys(")) | capture("^delete_matching_keys\\([a-z.]+, \"(?<re>.*)\"\\)$").re)] | unique) as $keys
          | (["authorization", "cookie", "session", "token", "header", "client", "forwarded", "question", "prompt", "user"][] as $w
              | if ($keys | length) == 0 or ($keys | map(contains($w)) | all | not)
                  then "the keys transform/scrub drops do not name \($w)" else empty end),
        ([($s | .. | strings | select(startswith("replace_all_patterns(")) | capture("^replace_all_patterns\\([a-z.]+, \"value\", \"(?<re>.*)\", \"\\[redacted\\]\"\\)$").re)] | unique) as $vals
          | (["[0-9]{1,3}", "ofsess", "ofenroll", "ofs", "bearer", "eyJ", "{40,}"][] as $w
              | if ($vals | length) == 0 or ($vals | map(contains($w)) | all | not)
                  then "the values transform/scrub rewrites do not name \($w)" else empty end)
      )' "$config" || echo "the collector configuration $config can not be read"
  gojq -r --yaml-input -s '
    (map(select(.kind == "ConfigMap" and (.metadata.name | startswith("otel-collector-config")))) | .[0]) as $cm
    | (map(select(.kind == "DaemonSet" and .metadata.name == "otel-collector")) | .[0]) as $ds
    | (map(select(.kind == "NetworkPolicy" and .metadata.name == "otel-collector")) | .[0]) as $np
    | ($ds.spec.template.spec // {}) as $pod
    | ($pod.containers[0] // {}) as $ct
    | (
        (if map(select(.kind == "Secret")) | length > 0 then "the telemetry layer holds a Secret: grafana-otlp is made out of band, never in git" else empty end),
        (if $ds == null then "the telemetry layer has no DaemonSet otel-collector" else empty end),
        (if $np == null then "the telemetry layer has no NetworkPolicy otel-collector" else empty end),
        (if ($ct.image // "") | test(":[A-Za-z0-9_][A-Za-z0-9_.-]*@sha256:[0-9a-f]{64}$") | not
          then "the collector image is not pinned by tag and digest, which Renovate moves together" else empty end),
        (if ($ct.securityContext.allowPrivilegeEscalation != false or $ct.securityContext.readOnlyRootFilesystem != true or ($ct.securityContext.capabilities.drop // []) != ["ALL"])
          then "the collector may escalate, write its root filesystem or hold a capability" else empty end),
        (if $pod.automountServiceAccountToken != false then "the collector pod mounts a Kubernetes API token" else empty end),
        ([$pod.volumes[]? | select(.hostPath != null) | .hostPath.path] | sort) as $hostPaths
          | (if $hostPaths != ["/var/lib/openfief-otel", "/var/log/pods"]
              then "the collector mounts hostPaths other than /var/log/pods and its own /var/lib/openfief-otel: \($hostPaths | join(" "))" else empty end),
        ([$pod.volumes[]? | select(.hostPath.path == "/var/log/pods") | .name][0]) as $logs
          | (if ([$ct.volumeMounts[]? | select(.name == $logs and .readOnly == true)] | length) == 0
              then "the collector mounts the node'"'"'s /var/log/pods writable" else empty end),
        ([$ct.env[]? | select(.valueFrom.secretKeyRef.name == "grafana-otlp" and .valueFrom.secretKeyRef.optional != true) | .valueFrom.secretKeyRef.key] | sort) as $keys
          | (if $keys != ["endpoint", "instance-id", "token"]
              then "the collector does not take endpoint, instance-id and token, all required, from the Secret grafana-otlp" else empty end),
        (if $cm == null or ($pod.volumes | map(select(.configMap.name == $cm.metadata.name)) | length) == 0
          then "the collector does not mount its ConfigMap" else empty end),
        ($np.spec // {}) as $sp
          | (if (($sp.policyTypes // []) | sort) != ["Egress", "Ingress"] then "the collector'"'"'s NetworkPolicy does not restrict both ingress and egress" else empty end),
            (if ($sp.ingress // []) | length == 0 or any(.[]; ([.ports[]?.port] | unique) != [4318] or (.from // [] | length) == 0 or any(.from[]; .namespaceSelector.matchLabels["kubernetes.io/metadata.name"] != "openfief"))
              then "the collector'"'"'s ingress is not the OTLP port 4318 from namespace openfief alone" else empty end),
            ($sp.egress // [] | .[]
              | (if (.ports // [] | length) == 0 or any(.ports[]; (.port != 53 and .port != 443))
                  then "the collector'"'"'s egress has a rule that is not DNS or HTTPS" else empty end),
                (if any(.ports[]?; .port == 443) and (.to // [] | any(.[]; .ipBlock == null))
                  then "the collector'"'"'s HTTPS egress goes to something other than an IP block" else empty end),
                (if any(.ports[]?; .port == 443) and (.to // [] | any(.[]; (.ipBlock.cidr == "0.0.0.0/0") and ((["10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16"] - (.ipBlock.except // [])) | length > 0)))
                  then "the collector'"'"'s HTTPS egress to the internet does not leave out the private, shared, link-local and loopback ranges" else empty end),
                (if any(.ports[]?; .port == 53) and (.to // [] | any(.[]; .podSelector.matchLabels["k8s-app"] != "kube-dns"))
                  then "the collector'"'"'s DNS egress is not to kube-dns" else empty end)),
            (if ($sp.egress // [] | map(.ports[]?.port) | unique) != [53, 443] then "the collector'"'"'s egress is not DNS and HTTPS" else empty end)
      )' "$2" | sed 's/^/the telemetry layer: /' || echo "the telemetry layer build can not be read"
  gojq -r --yaml-input -s '
    (map(select(.kind == "ConfigMap" and .metadata.name == "openfief-tracing")) | .[0]) as $cm
    | (map(select(.kind == "NetworkPolicy" and .metadata.name == "openfief-core-to-collector")) | .[0]) as $np
    | (
        (if ($cm.data.OTEL_EXPORTER_OTLP_ENDPOINT // "") != "http://otel-collector.observability:4318"
          then "the core and the operator do not send OTLP to the collector'"'"'s Service http://otel-collector.observability:4318" else empty end),
        (if $np == null or ($np.spec.egress | map(.ports[]?.port) | unique) != [4318]
            or ($np.spec.egress | map(.to[]?) | any(.namespaceSelector.matchLabels["kubernetes.io/metadata.name"] != "observability" or .podSelector.matchLabels["app.kubernetes.io/name"] != "otel-collector"))
          then "the core'"'"'s policy does not reach the collector'"'"'s pods in observability on port 4318 alone" else empty end)
      )' "$3" | sed 's/^/the demo: /' || echo "the live demo build can not be read"
  if ! cmp -s <(gojq -j --yaml-input 'select(.kind == "ConfigMap" and (.metadata.name | startswith("otel-collector-config"))) | .data["config.yaml"]' "$2") "$config"; then
    echo "the ConfigMap the collector mounts is not $config"
  fi
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
  # The telemetry layer is checked on its own build and on the live demo's (the core's endpoint and policy).
  live="$(mktemp)"
  if kustomize build "$dir/telemetry" >"$out" 2>/dev/null && kustomize build "$dir/openfief" >"$live" 2>/dev/null; then
    while IFS= read -r line; do
      [ -n "$line" ] || continue
      echo "demo: $line" >&2
      failures=$((failures + 1))
    done < <(telemetry_failures "$dir" "$out" "$live")
  else
    echo "demo: the telemetry layer ($dir/telemetry) or the live demo does not build" >&2
    failures=$((failures + 1))
  fi
  rm -f "$out" "$live"
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
