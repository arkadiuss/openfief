#!/usr/bin/env bash
# Builds the images the end-to-end suite runs from this checkout, so a pull request tests its own code instead of what
# GHCR holds, and makes them available to the local cluster in one of two ways:
#   - core, operator, egress-proxy and demo-seed are loaded into the nodes under the tags test/e2e/overlays/openfief
#     names, openfief.test/<name>:e2e (.test never resolves, so nothing pulls them);
#   - the plugin images the operator deploys, nbp-fx and the built-in UI, go to the suite's registry, because the
#     operator deploys a plugin only after it has read the image's signature from its registry (D-56, D-57). The suite
#     signs them there with a key it makes at run time (test/e2e/registry_test.go).
# The registry is a container, openfief-e2e-registry, on the cluster's Docker network, published on 127.0.0.1:5000 for
# the suite. The operator reaches it at its address on that network, a private IPv4 address, which the operator's
# registry client reads over plain HTTP. The nodes reach it there too, but containerd pulls only localhost over plain
# HTTP unless configured, so each plugin image is pulled into every node here with ctr, under the name the operator
# deploys; the operator asks for IfNotPresent, so the kubelet never pulls it itself.
# It writes the two registry addresses to <env file>, which task e2e passes to the suite.
# Usage: e2e-images.sh <k3d|kind> <cluster name> <env file>
set -euo pipefail
shopt -s inherit_errexit

usage='usage: e2e-images.sh <k3d|kind> <cluster name> <env file>'
cluster_kind="${1:?$usage}"
cluster_name="${2:?$usage}"
env_file="${3:?$usage}"
case "$cluster_kind" in
  k3d | kind) ;;
  *)
    echo "e2e-images: unknown cluster kind '$cluster_kind'; use k3d or kind" >&2
    exit 1
    ;;
esac

# registry:3.0.0 (distribution); the suite's signatures go to its referrers tag, since it has no Referrers API.
registry_image='docker.io/library/registry:3.0.0@sha256:6c5666b861f3505b116bb9aa9b25175e71210414bd010d92035ff64018f9457e'
registry_container=openfief-e2e-registry
registry_local=localhost:5000
# Every manifest type the registry may hold, so it answers with the stored manifest's own digest.
manifest_types='application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json'

images=()
plugins=()
for dockerfile in cmd/core/Dockerfile cmd/operator/Dockerfile cmd/egress-proxy/Dockerfile cmd/demo-seed/Dockerfile \
  plugins/nbp-fx/Dockerfile plugins/ui/Dockerfile; do
  name="$(basename "$(dirname "$dockerfile")")"
  image="openfief.test/$name:e2e"
  echo "e2e-images: building $image from $dockerfile"
  docker build --file "$dockerfile" --tag "$image" .
  case "$dockerfile" in
    plugins/*) plugins+=("$name") ;;
    *) images+=("$image") ;;
  esac
done

nodes=()
if [ "$cluster_kind" = k3d ]; then
  while IFS= read -r node; do nodes+=("$node"); done < <(docker ps --filter "label=k3d.cluster=$cluster_name" \
    --format '{{.Names}} {{.Label "k3d.role"}}' | awk '$2 == "server" || $2 == "agent" { print $1 }')
else
  while IFS= read -r node; do nodes+=("$node"); done < <(docker ps --filter "label=io.x-k8s.kind.cluster=$cluster_name" \
    --format '{{.Names}}')
fi
if [ "${#nodes[@]}" -eq 0 ]; then
  echo "e2e-images: no node found for $cluster_kind cluster $cluster_name" >&2
  exit 1
fi

# missing prints "<node> <image>" for each image absent from a node's containerd.
missing() {
  local node image present
  for node in "${nodes[@]}"; do
    present="$(docker exec "$node" ctr --namespace k8s.io images list --quiet)"
    for image in "$@"; do
      grep -qx "$image" <<<"$present" || grep -qx "docker.io/$image" <<<"$present" || echo "$node $image"
    done
  done
}

echo "e2e-images: loading ${images[*]} into $cluster_kind cluster $cluster_name"
if [ "$cluster_kind" = k3d ]; then
  # k3d's default import goes through a tools node and a shared tarball; when that tarball is gone before ctr reads it,
  # k3d logs an error and still exits 0 with "Successfully imported", and every pod then fails with ErrImagePull.
  # Direct mode streams into each node, and the check below fails here instead of minutes later in the suite.
  for attempt in 1 2 3; do
    k3d image import --mode direct --cluster "$cluster_name" "${images[@]}"
    gaps="$(missing "${images[@]}")"
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

# cluster_network prints the Docker network of the cluster's first node.
cluster_network() {
  docker inspect --format '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' "${nodes[0]}" |
    head -n 1
}

# start_registry starts a fresh registry container on network $1 and prints its address there.
start_registry() {
  local network="$1" address
  docker rm --force "$registry_container" >/dev/null 2>&1 || true
  docker run --detach --name "$registry_container" --network "$network" --publish 127.0.0.1:5000:5000 \
    "$registry_image" >/dev/null
  address="$(docker inspect --format "{{(index .NetworkSettings.Networks \"$network\").IPAddress}}" "$registry_container")"
  if [ -z "$address" ]; then
    echo "e2e-images: the registry has no address on network $network" >&2
    return 1
  fi
  for _ in $(seq 30); do
    curl -fsS -o /dev/null "http://$registry_local/v2/" && break
    sleep 1
  done
  curl -fsS -o /dev/null "http://$registry_local/v2/"
  echo "$address:5000"
}

# manifest_digest prints the digest of the manifest the registry holds for tag e2e of repository $1.
manifest_digest() {
  curl -fsS --head --header "Accept: $manifest_types" "http://$registry_local/v2/$1/manifests/e2e" |
    tr -d '\r' | awk 'tolower($1) == "docker-content-digest:" { print $2 }'
}

network="$(cluster_network)"
echo "e2e-images: starting the registry on Docker network $network"
cluster_registry="$(start_registry "$network")"
for name in "${plugins[@]}"; do
  docker tag "openfief.test/$name:e2e" "$registry_local/$name:e2e"
  docker push --quiet "$registry_local/$name:e2e"
  digest="$(manifest_digest "$name")"
  if [ -z "$digest" ]; then
    echo "e2e-images: the registry reports no digest for $name:e2e" >&2
    exit 1
  fi
  ref="$cluster_registry/$name@$digest"
  echo "e2e-images: pulling $ref into the nodes"
  for node in "${nodes[@]}"; do
    docker exec "$node" ctr --namespace k8s.io images pull --plain-http "$ref" >/dev/null
  done
  gaps="$(missing "$ref")"
  if [ -n "$gaps" ]; then
    echo "e2e-images: $ref is missing from nodes after the pull:" >&2
    echo "$gaps" >&2
    exit 1
  fi
done

mkdir -p "$(dirname "$env_file")"
printf 'export OPENFIEF_E2E_REGISTRY=%s\nexport OPENFIEF_E2E_CLUSTER_REGISTRY=%s\n' "$registry_local" "$cluster_registry" >"$env_file"
echo "e2e-images: the suite's registry is $registry_local, $cluster_registry in the cluster ($env_file)"
