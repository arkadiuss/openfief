#!/usr/bin/env bash
# Gives the local end-to-end cluster a Gateway API implementation, which the suite's exposure test attaches the
# HTTPRoute component (deploy/gateway) to and reaches the core through over HTTPS (test/e2e/exposure_test.go).
# OpenFief ships no Gateway: the owner brings one (D-28), so the suite brings one per cluster, as an owner might:
#   - kind: Envoy Gateway, from its release manifest pinned by version and sha256 (which also installs the Gateway
#     API CRDs), and the GatewayClass openfief-e2e for its controller;
#   - k3d: k3s's bundled Traefik with its Gateway API provider, which the CI job turns on when it creates the cluster
#     (test/e2e/testdata/traefik-gateway.yaml in k3s's manifests directory), and its GatewayClass traefik. When k3s's
#     Traefik CRD chart brought no Gateway API CRDs, the standard ones are installed from their release manifest,
#     pinned by version and sha256.
# Downloads go to <download dir> and are checked against their sha256 before anything is applied.
# Usage: e2e-gateway.sh <k3d|kind> <envoy gateway version> <its install.yaml sha256> <gateway api version>
#        <its standard-install.yaml sha256> <download dir>
set -euo pipefail
shopt -s inherit_errexit

usage='usage: e2e-gateway.sh <k3d|kind> <envoy gateway version> <sha256> <gateway api version> <sha256> <download dir>'
cluster_kind="${1:?$usage}"
envoy_version="${2:?$usage}"
envoy_sha256="${3:?$usage}"
gateway_api_version="${4:?$usage}"
gateway_api_sha256="${5:?$usage}"
download_dir="${6:?$usage}"

# How long each wait may take.
wait_timeout=5m
# The suite's GatewayClass on kind (test/e2e/exposure_test.go names it).
envoy_class=openfief-e2e

# fetch downloads url to file, unless the file is there already, and checks it against sha256.
fetch() {
  local url="$1" file="$2" sha256="$3"
  mkdir -p "$(dirname "$file")"
  if [ ! -f "$file" ]; then
    echo "e2e-gateway: downloading $url"
    bash "$(dirname "$0")/retry.sh" curl -fsSL -o "$file.part" "$url"
    mv "$file.part" "$file"
  fi
  if ! echo "$sha256  $file" | sha256sum -c - >/dev/null; then
    echo "e2e-gateway: $file does not match its pinned sha256 $sha256" >&2
    rm -f "$file"
    return 1
  fi
}

# gateway_api_crds waits for k3s's Traefik CRD chart and installs the pinned standard Gateway API CRDs when it
# brought none; it prints "installed" then, and its progress on stderr.
gateway_api_crds() {
  local crd=gateways.gateway.networking.k8s.io
  echo "e2e-gateway: waiting for k3s's Traefik CRD chart" >&2
  kubectl -n kube-system wait --for=condition=complete job/helm-install-traefik-crd --timeout="$wait_timeout" >&2
  if kubectl get crd "$crd" >/dev/null 2>&1; then
    echo "e2e-gateway: k3s's Traefik CRD chart brought the Gateway API CRDs" >&2
    return 0
  fi
  local file="$download_dir/gateway-api-$gateway_api_version-standard-install.yaml"
  fetch "https://github.com/kubernetes-sigs/gateway-api/releases/download/$gateway_api_version/standard-install.yaml" \
    "$file" "$gateway_api_sha256" >&2
  echo "e2e-gateway: installing the Gateway API $gateway_api_version standard CRDs" >&2
  kubectl apply --server-side -f "$file" >&2
  kubectl wait --for=condition=established "crd/$crd" --timeout="$wait_timeout" >&2
  echo installed
}

case "$cluster_kind" in
  kind)
    file="$download_dir/envoy-gateway-$envoy_version-install.yaml"
    fetch "https://github.com/envoyproxy/gateway/releases/download/$envoy_version/install.yaml" "$file" "$envoy_sha256"
    echo "e2e-gateway: installing Envoy Gateway $envoy_version"
    kubectl apply --server-side -f "$file"
    kubectl -n envoy-gateway-system wait --for=condition=available deployment/envoy-gateway --timeout="$wait_timeout"
    kubectl apply --server-side -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: $envoy_class
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
EOF
    kubectl wait --for=condition=accepted "gatewayclass/$envoy_class" --timeout="$wait_timeout"
    ;;
  k3d)
    crds="$(gateway_api_crds)"
    echo "e2e-gateway: waiting for Traefik with its Gateway API provider"
    kubectl -n kube-system wait --for=condition=complete job/helm-install-traefik --timeout="$wait_timeout"
    if [ "$crds" = installed ]; then
      # Traefik started before the CRDs existed: restart it, so its Gateway API provider watches them.
      kubectl -n kube-system rollout restart deployment/traefik
    fi
    kubectl -n kube-system rollout status deployment/traefik --timeout="$wait_timeout"
    kubectl wait --for=condition=accepted gatewayclass/traefik --timeout="$wait_timeout"
    ;;
  *)
    echo "e2e-gateway: unknown cluster kind '$cluster_kind'; use k3d or kind" >&2
    exit 1
    ;;
esac
echo "e2e-gateway: the cluster's Gateway API implementation is ready"
