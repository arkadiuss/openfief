#!/usr/bin/env bash
# Builds the openfief CLI for every release (docs/dev/sdlc.md#releases): Linux and macOS on amd64 and arm64,
# static, reproducible paths, with one CycloneDX SBOM (the dependencies are the same on every platform) and
# a SHA256SUMS file. Signing is a separate step (task release:cli:sign, CI only).
# Usage: release-cli.sh <version> [out dir]; needs cyclonedx-gomod on PATH unless SBOM=0.
set -euo pipefail

version="${1:?version}"
out="${2:-dist/cli}"
mkdir -p "$out"

for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  os="${platform%/*}"
  arch="${platform#*/}"
  bin="$out/openfief-$version-$os-$arch"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -buildvcs=true -ldflags='-s -w' -o "$bin" ./cmd/openfief
  echo "release-cli: built $bin"
done

if [ "${SBOM:-1}" = 1 ]; then
  cyclonedx-gomod app -json -licenses -main cmd/openfief -output "$out/openfief-$version.cdx.json" .
  echo "release-cli: wrote $out/openfief-$version.cdx.json"
fi

rm -f "$out/openfief-$version-SHA256SUMS"
(cd "$out" && sha256sum openfief-"$version"[-.]* >"openfief-$version-SHA256SUMS")
echo "release-cli: wrote $out/openfief-$version-SHA256SUMS"
