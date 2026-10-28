#!/usr/bin/env bash
# Lists the container images CI builds, as a JSON array for a GitHub Actions matrix (D-17).
# An image is a cmd/<name>/ or plugins/<name>/ directory with a Dockerfile; the build context is the
# repository root. Prints [] when there are none, so the image jobs are skipped rather than failed.
# Usage: ci-images.sh [registry prefix, e.g. ghcr.io/owner/openfief]
set -euo pipefail

# images prints one matrix entry per Dockerfile: its name, Dockerfile and lowercased image reference.
images() {
  local prefix dockerfile name first=1 seen=" "
  prefix="$(tr '[:upper:]' '[:lower:]' <<<"${1:-openfief}")"
  printf '['
  for dockerfile in cmd/*/Dockerfile plugins/*/Dockerfile; do
    [ -f "$dockerfile" ] || continue
    name="$(basename "$(dirname "$dockerfile")")"
    if [[ ! "$name" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
      echo "ci-images: $dockerfile: image name '$name' must be lowercase letters, digits and dashes" >&2
      return 1
    fi
    if [[ "$seen" == *" $name "* ]]; then
      echo "ci-images: $dockerfile: image name '$name' is used by another Dockerfile" >&2
      return 1
    fi
    seen+="$name "
    [ "$first" -eq 1 ] || printf ','
    printf '{"name":"%s","dockerfile":"%s","image":"%s/%s"}' "$name" "$dockerfile" "$prefix" "$name"
    first=0
  done
  printf ']\n'
}

images "${1:-}"
