#!/usr/bin/env bash
# Adds each first-party plugin to the catalog with the image digest the release signed (D-88), so first-party
# plugins show as verified: the author's hand-started release is the review D-44 requires.
# Usage: release-catalog.sh <version> <registry prefix> <digests dir> [catalog file]
#   <digests dir> holds one file per plugin image, named after it, with the sha256 digest images.yml pushed.
#   With VERIFY=1 (the release workflow) each digest's keyless signature by OpenFief's CI is checked with cosign first.
# Every plugins/<name>/Dockerfile needs a digest; a name and version already in the catalog is refused.
set -euo pipefail

version="${1:?version}"
prefix="$(tr '[:upper:]' '[:lower:]' <<<"${2:?registry prefix}")"
digests="${3:?digests dir}"
catalog="${4:-internal/plugins/catalog/plugins.json}"

# OpenFief's CI identity, as plugins.OpenFiefCI() trusts it (internal/plugins/trust.go).
issuer='https://token.actions.githubusercontent.com'
subject='https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/main'

entries='[]'
for dockerfile in plugins/*/Dockerfile; do
  [ -f "$dockerfile" ] || continue
  name="$(basename "$(dirname "$dockerfile")")"
  if [ ! -f "$digests/$name" ]; then
    echo "release-catalog: no signed digest for plugin $name in $digests" >&2
    exit 1
  fi
  digest="$(tr -d '[:space:]' <"$digests/$name")"
  if [[ ! "$digest" =~ ^sha256:[0-9a-f]{64}$ ]]; then
    echo "release-catalog: $digests/$name: '$digest' is not a sha256 digest" >&2
    exit 1
  fi
  image="$prefix/$name@$digest"
  if [ "${VERIFY:-0}" = 1 ]; then
    cosign verify --certificate-identity "$subject" --certificate-oidc-issuer "$issuer" "$image" >/dev/null
  fi
  if gojq -e --arg n "$name" --arg v "$version" 'any(.plugins[]?; .name == $n and .version == $v)' "$catalog" >/dev/null; then
    echo "release-catalog: $catalog already has $name $version" >&2
    exit 1
  fi
  entries="$(gojq -c --arg n "$name" --arg v "$version" --arg i "$image" --arg iss "$issuer" --arg sub "$subject" \
    '. + [{name: $n, version: $v, image: $i, publisher: {name: "openfief", issuer: $iss, subject: $sub}}]' <<<"$entries")"
done

tmp="$(mktemp)"
gojq --indent 2 --argjson add "$entries" '.plugins = ((.plugins // []) + $add)' "$catalog" >"$tmp"
mv "$tmp" "$catalog"
echo "release-catalog: added $(gojq 'length' <<<"$entries") plugins at $version to $catalog"
