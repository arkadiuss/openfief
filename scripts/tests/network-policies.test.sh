#!/usr/bin/env bash
# Builds the install's bases and checks the NetworkPolicies D-184 asks for: the database, the egress proxy and the
# operator each admit only their callers, and the egress proxy's own way out leaves every private, shared, link-local
# and loopback range out (kibi: TEST-deploy-network-policies).
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

for base in core egress-proxy operator demo-data; do
  kustomize build "$root/deploy/$base" >"$work/$base.yaml"
done

policy() { printf '[.[] | select(.kind == "NetworkPolicy" and .metadata.name == "%s")] | .[0]' "$1"; }

db="$(policy openfief-db)"
check "the database policy selects the CloudNativePG cluster openfief-db, ingress only" \
  "$db | .metadata.namespace == \"openfief\" and .spec.podSelector == {matchLabels: {\"cnpg.io/cluster\": \"openfief-db\"}} and .spec.policyTypes == [\"Ingress\"]" "$work/core.yaml"
check "the database admits the core on 5432" \
  "$db | .spec.ingress[0] == {from: [{podSelector: {matchLabels: {\"app.kubernetes.io/name\": \"openfief-core\"}}}], ports: [{protocol: \"TCP\", port: 5432}]}" "$work/core.yaml"
check "the database admits its own instances on 5432 and 8000" \
  "$db | .spec.ingress[1] == {from: [{podSelector: {matchLabels: {\"cnpg.io/cluster\": \"openfief-db\"}}}], ports: [{protocol: \"TCP\", port: 5432}, {protocol: \"TCP\", port: 8000}]}" "$work/core.yaml"
check "the database admits CloudNativePG's operator on 8000 alone" \
  "$db | .spec.ingress[2] == {from: [{namespaceSelector: {matchLabels: {\"kubernetes.io/metadata.name\": \"cnpg-system\"}}, podSelector: {matchLabels: {\"app.kubernetes.io/name\": \"cloudnative-pg\"}}}], ports: [{protocol: \"TCP\", port: 8000}]}" "$work/core.yaml"
check "the database admits nothing else, plugins included" \
  "$db | .spec.ingress | length == 3 and all(.[]; .from | all(.[]; (.podSelector.matchLabels[\"app.kubernetes.io/name\"] // \"\") != \"openfief-plugin\"))" "$work/core.yaml"
check "the demo's loader brings its own admission to the database" \
  "$(policy openfief-db-from-demo-seed) | .spec.podSelector.matchLabels[\"cnpg.io/cluster\"] == \"openfief-db\" and .spec.ingress == [{from: [{podSelector: {matchLabels: {\"app.kubernetes.io/name\": \"demo-seed\"}}}], ports: [{protocol: \"TCP\", port: 5432}]}]" "$work/demo-data.yaml"

proxy="$(policy egress-proxy)"
check "the egress proxy admits plugin pods on 4750 alone" \
  "$proxy | .metadata.namespace == \"openfief\" and .spec.policyTypes == [\"Ingress\", \"Egress\"] and .spec.ingress == [{from: [{podSelector: {matchLabels: {\"app.kubernetes.io/name\": \"openfief-plugin\"}}}], ports: [{protocol: \"TCP\", port: 4750}]}]" "$work/egress-proxy.yaml"
check "the egress proxy's way out leaves the private, shared, link-local and loopback ranges out" \
  "$proxy | .spec.egress[0] == {to: [{ipBlock: {cidr: \"0.0.0.0/0\", except: [\"10.0.0.0/8\", \"100.64.0.0/10\", \"127.0.0.0/8\", \"169.254.0.0/16\", \"172.16.0.0/12\", \"192.168.0.0/16\"]}}, {ipBlock: {cidr: \"::/0\", except: [\"::1/128\", \"fc00::/7\", \"fe80::/10\"]}}]}" "$work/egress-proxy.yaml"
check "the egress proxy reaches the API server on 6443 and DNS, and nothing else" \
  "$proxy | (.spec.egress | length == 3) and (.spec.egress[1].ports == [{protocol: \"TCP\", port: 6443}]) and (.spec.egress[2].to[0].podSelector.matchLabels[\"k8s-app\"] == \"kube-dns\")" "$work/egress-proxy.yaml"
check "the egress proxy is told its node's address" \
  '[.[] | select(.kind == "Deployment" and .metadata.name == "egress-proxy") | .spec.template.spec.containers[0].env[] | select(.name == "OPENFIEF_NODE_IP") | .valueFrom.fieldRef.fieldPath] == ["status.hostIP"]' "$work/egress-proxy.yaml"

check "the operator admits no ingress" \
  "$(policy openfief-operator) | .spec.podSelector == {matchLabels: {\"app.kubernetes.io/name\": \"openfief-operator\"}} and .spec.policyTypes == [\"Ingress\"] and (.spec.ingress // []) == []" "$work/operator.yaml"

if [ "$failures" -ne 0 ]; then
  echo "network-policies: $failures of $cases cases failed" >&2
  exit 1
fi
echo "network-policies: $cases cases passed"
