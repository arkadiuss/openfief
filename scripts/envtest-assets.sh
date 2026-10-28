#!/usr/bin/env bash
# Installs the Kubernetes API server and etcd that the operator's envtest suite runs against
# (internal/plugins/operator), from controller-tools' envtest releases, checked against pinned SHA-512 sums.
# Run once by `task tools`; tests then run offline with KUBEBUILDER_ASSETS pointing at the result.
# Usage: envtest-assets.sh <kubernetes version, e.g. 1.36.2> <directory>; installs into <directory>/<version>.
# The sums come from controller-tools' envtest-releases.yaml; bump them together with ENVTEST_K8S_VERSION.
set -euo pipefail

version="${1:?kubernetes version}"
dir="${2:?install directory}"

platform="$(go env GOOS)-$(go env GOARCH)"
# The pinned SHA-512 of the envtest archive for this platform.
case "$version/$platform" in
1.36.2/linux-amd64) want=ea743186c8a799f5cf8faf16969f86189d003cb7d130e0ac4b58789f1e5748dcf30ebe91c837a10d5ac415383da3e10b9e64d65785c938c23e739781cfb76f08 ;;
1.36.2/linux-arm64) want=2d72ee985a8e262a3c57dc9f7f0fd891f6a8c7bf7ebaa2db6dc6d8eac8ae28181afe51c1f368b67756cdb40b10de9b205609e1726e5f27f7c6d824dd9c6649ac ;;
1.36.2/darwin-amd64) want=b4fe9f973cd1992e3880f8b230ed32c0c243b83371591e3e3315f4404df2cdbef8ee94eeb9bb245b7e5feb987d5c4a76c43e0877654ab73f00d7bbc2b9984514 ;;
1.36.2/darwin-arm64) want=9278f9e5af556b2f1f2d139769c1f0d717c7b4426917fdebdba898bcb725a916e4910d6160886194ff9be9589ea7c5c32c2b8ae0754703874b7c8ba8ddfc41ce ;;
*)
  echo "envtest-assets: no pinned checksum for envtest $version on $platform" >&2
  exit 1
  ;;
esac
archive="envtest-v${version}-${platform}.tar.gz"
url="https://github.com/kubernetes-sigs/controller-tools/releases/download/envtest-v${version}/${archive}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -fsSL --retry 3 -o "$tmp/$archive" "$url"
if command -v sha512sum >/dev/null; then
  got="$(sha512sum "$tmp/$archive" | cut -d' ' -f1)"
else
  got="$(shasum -a 512 "$tmp/$archive" | cut -d' ' -f1)"
fi
if [ "$got" != "$want" ]; then
  echo "envtest-assets: $archive: SHA-512 $got does not match the pinned $want" >&2
  exit 1
fi

mkdir -p "$tmp/out"
tar -xzf "$tmp/$archive" -C "$tmp/out"
rm -rf "${dir:?}/$version"
mkdir -p "$dir"
mv "$tmp/out/controller-tools/envtest" "$dir/$version"
echo "envtest $version installed in $dir/$version"
