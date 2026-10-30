#!/usr/bin/env bash
# Builds the Gateway component (deploy/gateway, D-28) the way docs/guides/expose.md tells an owner to, with the
# guide's own overlay, and checks what it adds; checks that the product bases alone expose nothing (D-28) and that
# the core's ports are split as D-115 says (kibi: TEST-deploy-gateway-component).
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
     inside {block = block "\n" $0}' "$root/docs/guides/expose.md" >"$work/deploy/fief/kustomization.yaml"
if ! grep -q '^components:' "$work/deploy/fief/kustomization.yaml"; then
  echo "FAIL: docs/guides/expose.md has no overlay with components" >&2
  exit 1
fi

kustomize build "$work/deploy/fief" >"$work/fief.yaml"
kustomize build "$work/deploy/core" >"$work/core.yaml"

route='[.[] | select(.kind == "HTTPRoute")]'
check "one route, to the core's http port" \
  "$route | length == 1 and (.[0].spec.rules == [{backendRefs: [{name: \"openfief-core\", port: 8080}]}])" "$work/fief.yaml"
check "the route on the owner's Gateway and hostname" \
  "$route | .[0].spec | .parentRefs[0].name == \"home\" and .parentRefs[0].namespace == \"gateway\" and .hostnames == [\"fief.example.org\"]" "$work/fief.yaml"
check "port 8080 is the Service's http port" \
  '[.[] | select(.kind == "Service" and .metadata.name == "openfief-core") | .spec.ports[] | select(.port == 8080) | .name] == ["http"]' "$work/fief.yaml"
check "the Gateway's proxies admitted to the browser port alone" \
  '[.[] | select(.kind == "NetworkPolicy" and .metadata.name == "openfief-core-from-gateway") | .spec.ingress[] | .ports] == [[{protocol: "TCP", port: 8080}]]' "$work/fief.yaml"
check "the core allows the route's origin, and localhost still" \
  '[.[] | select(.kind == "Deployment" and .metadata.name == "openfief-core") | .spec.template.spec.containers[0].args[] | select(startswith("--auth-origin"))] == ["--auth-origin=http://localhost:8080", "--auth-origin=https://fief.example.org"]' "$work/fief.yaml"
check "the core base alone has no route and admits no proxy" \
  '[.[] | select(.kind == "HTTPRoute" or .kind == "Ingress" or .metadata.name == "openfief-core-from-gateway")] == []' "$work/core.yaml"
check "plugin pods reach only the plugin port" \
  '[.[] | select(.kind == "NetworkPolicy" and .metadata.name == "openfief-core") | .spec.ingress[] | .ports] == [[{protocol: "TCP", port: 8081}]]' "$work/core.yaml"
check "the connect port targets the plugin port" \
  '[.[] | select(.kind == "Service" and .metadata.name == "openfief-core") | .spec.ports[] | {name, targetPort}] == [{name: "http", targetPort: 8080}, {name: "connect", targetPort: 8081}]' "$work/core.yaml"

if [ "$failures" -ne 0 ]; then
  echo "gateway: $failures of $cases cases failed" >&2
  exit 1
fi
echo "gateway: $cases cases passed"
