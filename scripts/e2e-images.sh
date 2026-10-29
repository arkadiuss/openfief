#!/usr/bin/env bash
# Builds the images the end-to-end suite runs (core, operator, egress-proxy, demo-seed and the built-in UI plugin) from
# this checkout and loads them into the local cluster, so a pull request tests its own code instead of what GHCR holds.
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
for dockerfile in cmd/core/Dockerfile cmd/operator/Dockerfile cmd/egress-proxy/Dockerfile cmd/demo-seed/Dockerfile \
  plugins/ui/Dockerfile; do
  name="$(basename "$(dirname "$dockerfile")")"
  image="openfief.test/$name:e2e"
  echo "e2e-images: building $image from $dockerfile"
  docker build --file "$dockerfile" --tag "$image" .
  images+=("$image")
done

# missing prints "<node> <image>" for each image absent from a k3d node's containerd.
missing() {
  local node image present
  for node in "${nodes[@]}"; do
    present="$(docker exec "$node" ctr --namespace k8s.io images list --quiet)"
    for image in "${images[@]}"; do
      grep -qx "$image" <<<"$present" || grep -qx "docker.io/$image" <<<"$present" || echo "$node $image"
    done
  done
}

echo "e2e-images: loading ${images[*]} into $cluster_kind cluster $cluster_name"
if [ "$cluster_kind" = k3d ]; then
  # k3d's default import goes through a tools node and a shared tarball; when that tarball is gone before ctr reads it,
  # k3d logs an error and still exits 0 with "Successfully imported", and every pod then fails with ErrImagePull.
  # Direct mode streams into each node, and the check below fails here instead of minutes later in the suite.
  nodes=()
  while IFS= read -r node; do nodes+=("$node"); done < <(docker ps --filter "label=k3d.cluster=$cluster_name" \
    --format '{{.Names}} {{.Label "k3d.role"}}' | awk '$2 == "server" || $2 == "agent" { print $1 }')
  if [ "${#nodes[@]}" -eq 0 ]; then
    echo "e2e-images: no server or agent node found for k3d cluster $cluster_name" >&2
    exit 1
  fi
  for attempt in 1 2 3; do
    k3d image import --mode direct --cluster "$cluster_name" "${images[@]}"
    gaps="$(missing)"
    [ -z "$gaps" ] && break
    echo "e2e-images: attempt $attempt left images missing:" >&2
    echo "$gaps" >&2
  done
  if [ -n "$gaps" ]; then
    echo "e2e-images: images still missing from the k3d nodes after 3 imports" >&2
    exit 1
  fi
else
  kind load docker-image --name "$cluster_name" "${images[@]}"
fi
