#!/usr/bin/env bash
# Tests for scripts/demo/validate.sh on copies of deploy/ (kibi: TEST-tooling-demo-validate).
# Needs kustomize and gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# fresh copies deploy/ into a throwaway repository root, so Flux paths (./deploy/...) resolve the same way.
fresh() {
  rm -rf "$work/repo"
  mkdir -p "$work/repo"
  cp -r "$root/deploy" "$work/repo/deploy"
}

# validate runs validate.sh in the throwaway root; output lands in $work/out.
validate() {
  (cd "$work/repo" && bash "$root/scripts/demo/validate.sh") >"$work/out" 2>&1
}

# expect records whether validate passed or failed as wanted, and that the output names the problem.
expect() {
  local name="$1" want="$2" pattern="$3" got=pass
  cases=$((cases + 1))
  validate || got=fail
  if [ "$got" != "$want" ] || ! grep -q "$pattern" "$work/out"; then
    echo "FAIL: $name: want $want with '$pattern', got $got:" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

fresh
expect "the committed manifests" pass "every image is pinned by digest"

fresh
sed -i 's/    provider: cosign/    provider: notation/' "$work/repo/deploy/demo/stack.yaml"
expect "an OCI source without cosign verification" fail "OCI sources without cosign verification: openfief-demo-app"

fresh
sed -i '/^images:/,$d' "$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "an image pinned by tag only" fail "not pinned by digest: ghcr.io/arkadiuss/openfief/"

fresh
sed -i 's#path: ./deploy/demo/database#path: ./deploy/demo/missing#' "$work/repo/deploy/demo/stack.yaml"
expect "a Flux path that does not build" fail "kustomize build ./deploy/demo/missing failed"

fresh
sed -i 's/[$]{DEMO_HOST}/${DEMO_HOSTNAME}/' "$work/repo/deploy/demo/routes/openfief.yaml"
expect "a variable the demo settings don't define" fail "uses variables the demo settings don't define: DEMO_HOSTNAME"

fresh
sed -i 's/  secrets: \[openrouter--api-key, openrouter--ca-bundle\]/  secrets: [openrouter--api-key]/' "$work/repo/deploy/demo/model/openrouter.yaml"
expect "a grant that leaves out a Secret its Plugin's manifest asks for" fail "PluginGrant openrouter does not approve the Secret openrouter--ca-bundle"

# The model layer (deploy/demo/model) is a Component the live demo includes; its switch is the owner's, not the manifests'.
fresh
sed -i 's#"openrouter/openai/gpt-oss-120b"#"openrouter/google/gemma-4-26b-a4b-it"#' "$work/repo/deploy/demo/model/settings.json"
expect "a route to a model the provider instance does not list" fail "name openrouter/google/gemma-4-26b-a4b-it, which the instance openrouter does not list"

fresh
sed -i 's#"routes": {#"routes": {\n    "maintainer": ["openrouter/openai/gpt-oss-120b"],#' "$work/repo/deploy/demo/model/settings.json"
expect "a route for the maintainer" fail "route something other than the Ask box"

fresh
sed -i '/^components:/,/^  - ..\/..\/tracing/d' "$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "the model layer left out of the live demo" fail "the core's flags lack --ask-questions-per-hour=10"

fresh
rm "$work/repo/deploy/demo/model/openrouter.yaml"
expect "a model layer that does not build" fail "the model layer does not build with the demo"

# D-208's guards in the model layer: the core's flags, the UI's config and its grant.
fresh
sed -i 's/value: --ask-concurrent-calls=3/value: --ask-concurrent-calls=30/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "calls at once that are not D-208's" fail "the core's flags lack --ask-concurrent-calls=3"

fresh
sed -i 's/value: --ask-questions-per-hour=10/value: --ask-questions-per-hour=100/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "an hourly limit that is not D-208's" fail "the core's flags lack --ask-questions-per-hour=10"

fresh
sed -i 's/value: --ask-questions-per-day=30/value: --ask-questions-per-day=300/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a daily limit that is not D-208's" fail "the core's flags lack --ask-questions-per-day=30"

fresh
sed -i 's/value: --allow-turnstile/value: --allow-turnstile-not/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "no place for Turnstile in the core's policy" fail "the core's flags lack --allow-turnstile"

fresh
sed -i 's/value: --trusted-proxies=10.42.0.0\/16/value: --trusted-proxies-not/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "no trusted proxy for the address header" fail "the core's flags lack --trusted-proxies"

fresh
sed -i 's/value: --model-daily-budget-usd=0.25/value: --model-daily-budget-usd=lots/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a budget that is no amount" fail "the core's --model-daily-budget-usd is not a positive amount"

fresh
sed -i 's/value: --model-daily-budget-usd=0.25/value: --model-daily-budget-usd=0/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a budget of nothing" fail "the core's --model-daily-budget-usd is not a positive amount"

fresh
sed -i 's/value: \[challenges.cloudflare.com\]/value: [elsewhere.example.com]/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a grant that does not approve Turnstile's host" fail "the UI grant does not approve the host of turnstileVerifyUrl"

fresh
sed -i 's/value: \[ui--turnstile-secret\]/value: [ui--other]/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a grant that does not name the Turnstile Secret" fail "the UI grant does not name the Secret ui--turnstile-secret"

fresh
sed -i 's/askOutputTokens: 800/askOutputTokens: 1600/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "limits of one question that are not D-208's" fail "the UI's limits of one question are not D-208's"

fresh
sed -i 's/turnstileSiteKey: .*/turnstileSiteKey: REPLACE-WITH-THE-TURNSTILE-SITE-KEY/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "the placeholder site key in the live demo" fail "has the placeholder Turnstile site key"

fresh
sed -i 's/turnstileSiteKey: .*/turnstileSiteKey: 0x4AAAAAAAxxxxxxxxxxxxxx/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a model layer with a site key" pass "every image is pinned by digest"

# The telemetry layer (deploy/demo/telemetry, TEST-demo-telemetry-scrub): the collector ships the demo's logs, traces and metrics
# to Grafana Cloud, so each way its scrub, its one exporter, its credentials or its network could let personal data out,
# or send it somewhere else, fails naming the problem.
tel="$work/repo/deploy/demo/telemetry"

fresh
sed -i '/name: demo-telemetry/,/^---/d' "$work/repo/deploy/demo/stack.yaml"
expect "the telemetry layer left out of the Flux stack" fail "does not apply ./deploy/demo/telemetry"

fresh
sed -i '/^    metrics:$/,/^      exporters/ s/processors: \[memory_limiter, transform\/scrub, transform\/service, batch\]/processors: [memory_limiter, transform\/service, batch]/' "$tel/collector.yaml"
expect "the scrub missing from the metrics pipeline" fail "the pipeline metrics does not run the processor transform/scrub"

fresh
sed -i '/^    logs:$/,/^      exporters/ s/processors: \[memory_limiter, transform\/scrub, transform\/service, batch\]/processors: [batch, transform\/scrub, transform\/service]/' "$tel/collector.yaml"
expect "a processor before the scrub in the logs pipeline" fail "the pipeline logs runs a processor other than the memory limiter before transform/scrub"

fresh
sed -i '/^    traces:$/,/^      exporters/ s/exporters: \[otlphttp\/grafana\]/exporters: [otlphttp\/grafana, debug]/' "$tel/collector.yaml"
expect "a second exporter in the traces pipeline" fail "the pipeline traces exports somewhere other than the one exporter"

fresh
sed -i 's/^    error_mode: propagate/    error_mode: ignore/' "$tel/collector.yaml"
expect "a scrub that fails open" fail "transform/scrub is not fail-closed"

fresh
sed -i '/delete_matching_keys(span.attributes/d' "$tel/collector.yaml"
expect "no key scrub of span attributes" fail "no delete_matching_keys of span.attributes in its trace_statements"

fresh
sed -i '/replace_all_patterns(datapoint.attributes/d' "$tel/collector.yaml"
expect "no value scrub of metric data points" fail "no replace_all_patterns of datapoint.attributes values in its metric_statements"

fresh
sed -i '/replace_pattern(log.body/d' "$tel/collector.yaml"
expect "the log body left unscrubbed" fail "does not rewrite the log body"

fresh
sed -i '/replace_pattern(span.status.message/d' "$tel/collector.yaml"
expect "a span's status message left unscrubbed" fail "does not rewrite span.status.message"

fresh
sed -i 's/authorization|//' "$tel/collector.yaml"
expect "Authorization not among the keys dropped" fail "the keys transform/scrub drops do not name authorization"

fresh
sed -i 's/question|questions|q|prompt|prompts|query|answer|completion/query|answer/' "$tel/collector.yaml"
expect "question and prompt not among the keys dropped" fail "the keys transform/scrub drops do not name question"

fresh
sed -i 's/(ofs|ofsess|ofenroll)_\[A-Za-z0-9_-\]+|//' "$tel/collector.yaml"
expect "the token prefixes not among the values rewritten" fail "the values transform/scrub rewrites do not name ofsess"

fresh
sed -i 's/^    endpoint: ${env:GRAFANA_OTLP_ENDPOINT}/    endpoint: https:\/\/otlp.example.com\/otlp/' "$tel/collector.yaml"
expect "a literal endpoint" fail "does not take its endpoint from"

fresh
sed -i 's/password: ${env:GRAFANA_OTLP_TOKEN}/password: glc_literal/' "$tel/collector.yaml"
expect "a literal token" fail "basicauth/grafana does not take its username and password from"

fresh
sed -i 's|^  # The Grafana Cloud OTLP gateway,|  # The gateway https://otlp-gateway-prod-eu-west-0.grafana.net/otlp,|' "$tel/collector.yaml"
expect "a Grafana address in git" fail "names a Grafana address"

fresh
printf 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: grafana-otlp\nstringData:\n  token: x\n' >"$tel/secret.yaml"
sed -i 's/^  - daemonset.yaml/  - daemonset.yaml\n  - secret.yaml/' "$tel/kustomization.yaml"
expect "a Secret in git" fail "the telemetry layer holds a Secret"

fresh
sed -i '0,/          port: 443/ s//          port: 8080/' "$tel/networkpolicy.yaml"
expect "egress to a port that is not HTTPS" fail "the telemetry layer: the collector's egress has a rule that is not DNS or HTTPS"

fresh
sed -i '/^              - 10.0.0.0\/8$/d' "$tel/networkpolicy.yaml"
expect "egress to the internet that leaves a private range in" fail "does not leave out the private, shared, link-local and loopback ranges"

fresh
sed -i '0,/          port: 4318/ s//          port: 4317/' "$tel/networkpolicy.yaml"
expect "an ingress port that is not OTLP/HTTP" fail "ingress is not the OTLP port 4318 from namespace openfief alone"

fresh
sed -i '/mountPath: \/var\/log\/pods/{n;d}' "$tel/daemonset.yaml"
expect "the node's pod logs mounted writable" fail "mounts the node's /var/log/pods writable"

fresh
sed -i 's#^            path: /var/lib/openfief-otel#            path: /var/lib#' "$tel/daemonset.yaml"
expect "a hostPath that is not the collector's own" fail "mounts hostPaths other than /var/log/pods"

fresh
sed -i '0,/optional: false/ s//optional: true/' "$tel/daemonset.yaml"
expect "a credential the pod does not wait for" fail "does not take endpoint, instance-id and token, all required, from the Secret grafana-otlp"

fresh
sed -i 's/drop: \[ALL\]/drop: []/' "$tel/daemonset.yaml"
expect "a collector that keeps its capabilities" fail "may escalate, write its root filesystem or hold a capability"

fresh
sed -i 's#:0.164.0@sha256:[0-9a-f]*#:0.164.0@sha256:0000#' "$tel/daemonset.yaml"
expect "a collector image without a full digest" fail "the collector image is not pinned by tag and digest"

fresh
sed -i 's#http://otel-collector.observability:4318#http://otel-collector.observability:4317#' "$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "the core sending somewhere other than the collector" fail "the demo: the core and the operator do not send OTLP to the collector's Service"

if [ "$failures" -ne 0 ]; then
  echo "demo-validate: $failures of $cases cases failed" >&2
  exit 1
fi
echo "demo-validate: $cases cases passed"
