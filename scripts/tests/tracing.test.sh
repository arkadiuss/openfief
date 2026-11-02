#!/usr/bin/env bash
# Builds the tracing component (deploy/tracing) the way docs/guides/tracing.md tells an owner to, with the guide's own
# overlay, and checks what it adds; checks that the product bases alone turn no tracing on and let the core reach no
# collector (kibi: TEST-deploy-tracing-component).
# Needs kustomize and gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# check records one case: the gojq filter $2 over the built manifests $3 must print true.
check() {
  local name="$1" filter="$2" file="$3" got
  cases=$((cases + 1))
  got="$(gojq --yaml-input -s "$filter" "$file" 2>&1)" || true
  if [ "$got" != "true" ]; then
    echo "FAIL: $name: $filter printed $got" >&2
    failures=$((failures + 1))
  fi
}

cp -r "$root/deploy" "$work/deploy"
mkdir -p "$work/deploy/fief"
# The guide's overlay: the first YAML block that lists components.
awk '/^```yaml$/ {block = ""; inside = 1; next}
     /^```$/ && inside {if (block ~ /\ncomponents:/) {printf "%s", block; exit} inside = 0; next}
     inside {block = block "\n" $0}' "$root/docs/guides/tracing.md" >"$work/deploy/fief/kustomization.yaml"
if ! grep -q '^components:' "$work/deploy/fief/kustomization.yaml"; then
  echo "FAIL: docs/guides/tracing.md has no overlay with components" >&2
  exit 1
fi

kustomize build "$work/deploy/fief" >"$work/fief.yaml"
kustomize build "$work/deploy/core" >"$work/core.yaml"
kustomize build "$work/deploy/operator" >"$work/operator.yaml"

container='.spec.template.spec.containers[0]'
check "the ConfigMap names the owner's collector and sampling" \
  '[.[] | select(.kind == "ConfigMap" and .metadata.name == "openfief-tracing") | .data] == [{OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.monitoring:4318", OTEL_TRACES_SAMPLER: "parentbased_traceidratio", OTEL_TRACES_SAMPLER_ARG: "0.25"}]' "$work/fief.yaml"
for deployment in openfief-core openfief-operator; do
  check "$deployment reads the ConfigMap" \
    "[.[] | select(.kind == \"Deployment\" and .metadata.name == \"$deployment\") | $container.envFrom] == [[{configMapRef: {name: \"openfief-tracing\"}}]]" "$work/fief.yaml"
  check "$deployment takes the headers from the optional Secret" \
    "[.[] | select(.kind == \"Deployment\" and .metadata.name == \"$deployment\") | $container.env[] | select(.name == \"OTEL_EXPORTER_OTLP_HEADERS\") | .valueFrom.secretKeyRef] == [{name: \"openfief-tracing\", key: \"headers\", optional: true}]" "$work/fief.yaml"
done
check "the core keeps its own settings" \
  "[.[] | select(.kind == \"Deployment\" and .metadata.name == \"openfief-core\") | $container.env[] | .name] == [\"OPENFIEF_DSN\", \"OPENFIEF_AUTH_ORIGINS\", \"OTEL_EXPORTER_OTLP_HEADERS\"]" "$work/fief.yaml"
check "the core reaches the owner's collector pods on the OTLP ports alone" \
  '[.[] | select(.kind == "NetworkPolicy" and .metadata.name == "openfief-core-to-collector") | .spec | {policyTypes, egress}] == [{policyTypes: ["Egress"], egress: [{to: [{namespaceSelector: {matchLabels: {"kubernetes.io/metadata.name": "monitoring"}}, podSelector: {matchLabels: {"app.kubernetes.io/name": "otel-collector"}}}], ports: [{protocol: "TCP", port: 4317}, {protocol: "TCP", port: 4318}]}]}]' "$work/fief.yaml"
check "the core base sets no tracing variable" \
  "[.[] | select(.kind == \"Deployment\") | $container | (.envFrom // []), [(.env // [])[] | select(.name | startswith(\"OTEL_\"))]] == [[], []]" "$work/core.yaml"
check "the operator base sets no tracing variable" \
  "[.[] | select(.kind == \"Deployment\") | $container | (.envFrom // []), (.env // [])] == [[], []]" "$work/operator.yaml"
check "the core base reaches no collector" \
  '[.[] | select(.kind == "NetworkPolicy") | .spec.egress[]?.ports[]?.port | select(. == 4317 or . == 4318)] == []' "$work/core.yaml"

if [ "$failures" -ne 0 ]; then
  echo "tracing: $failures of $cases cases failed" >&2
  exit 1
fi
echo "tracing: $cases cases passed"
