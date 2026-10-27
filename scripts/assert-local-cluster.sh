#!/usr/bin/env bash
# Refuses to run end-to-end tests against anything but a local cluster (docs/dev/testing.md).
set -euo pipefail

if ! command -v kubectl >/dev/null 2>&1; then
  echo "e2e: kubectl not found; start a local k3s (native or k3d) first" >&2
  exit 1
fi

context="$(kubectl config current-context 2>/dev/null || true)"
server="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' 2>/dev/null || true)"

case "$server" in
  https://127.0.0.1:* | https://localhost:* | https://0.0.0.0:* | https://\[::1\]:*)
    echo "e2e: target is local context '$context' ($server)"
    ;;
  *)
    echo "e2e: refusing context '${context:-none}' at '${server:-unknown}': not a local cluster" >&2
    exit 1
    ;;
esac
