#!/usr/bin/env bash
# Builds the images the end-to-end suite runs (core, operator, egress-proxy, demo-seed) from this checkout and
# loads them into the local cluster, so a pull request tests its own code instead of what GHCR holds.
# The tags match test/e2e/overlays/openfief: openfief.test/<name>:e2e (.test never resolves, so nothing pulls them).
# Usage: e2e-images.sh <k3d|kind> <cluster name>
set -euo pipefail

cluster_kind="${1:?usage: e2e-images.sh <k3d|kind> <cluster name>}"
cluster_name="${2:?usage: e2e-images.sh <k3d|kind> <cluster name>}"
case "$cluster_kind" in
  k3d | kind) ;;
  *)
    echo "e2e-images: unknown cluster kind '$cluster_kind'; use k3d or kind" >&2
    exit 1
    ;;
esac

images=()
for name in core operator egress-proxy demo-seed; do
  image="openfief.test/$name:e2e"
  echo "e2e-images: building $image from cmd/$name/Dockerfile"
  docker build --file "cmd/$name/Dockerfile" --tag "$image" .
  images+=("$image")
done

echo "e2e-images: loading ${images[*]} into $cluster_kind cluster $cluster_name"
if [ "$cluster_kind" = k3d ]; then
  k3d image import --cluster "$cluster_name" "${images[@]}"
else
  kind load docker-image --name "$cluster_name" "${images[@]}"
fi
