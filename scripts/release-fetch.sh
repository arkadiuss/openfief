#!/usr/bin/env bash
# Fetches a PUBLISHED release and verifies it the way the install guide tells the owner to (docs/guides/install.md,
# steps 1 to 3 and the checksum of step 4), without touching a cluster: the end-to-end suite's upgrade run installs
# it first and upgrades it to this checkout's bundle (task e2e:upgrade, test/e2e/upgrade_test.go). It downloads the
# release's own assets from its GitHub release, never anything this checkout builds, into an empty directory:
#   <dir>/openfief-{crds,install,db}-<version>.yaml, <dir>/cnpg-<cnpg version>.yaml   the bundle
#   <dir>/cli/                                                                         the CLI of every platform, the
#                                                                                      SBOM, checksums and signatures
#   <dir>/bin/openfief                                                                 the verified CLI for OS/ARCH
#   <dir>/demo-seed.image                                                              the release's demo-seed image,
#                                                                                      by the digest cosign verified
# and fails on the first thing that doesn't hold:
#   1. downloads the bundle and the CLI files with gh (retried on a server error, never on a client one);
#   2. verifies the CLI, the SBOM and the checksums file with `release-verify.sh cli`, which runs the guide's
#      cosign verify-blob as release.yml on main signed them and checks the checksums, and checks that cosign REFUSES
#      the CLI under a wrong identity (images.yml's), so the identity flags are not ignored;
#   3. verifies every image the install file runs with `release-verify.sh images`: each of ours pinned by digest and
#      signed by images.yml on main, SpiceDB's pinned by tag and digest;
#   4. checks CloudNativePG's manifest against its pinned sha256;
#   5. resolves the release's demo-seed image (<repo images prefix>/demo-seed:<version>) to the digest cosign
#      verifies as images.yml on main signed it (verified-digest.sh), and records it by digest: the upgrade run seeds
#      the install with the release's own loader, so the data is stored by that release's code.
# The release's CLI isn't run here.
#
# Usage: CNPG_VERSION=<x.y.z> CNPG_SHA256=<sha256> release-fetch.sh <version> <owner/repo> <empty dir>
# Optional: OS and ARCH of the CLI to install under bin/ (default linux, amd64). Needs gh (GH_TOKEN), cosign, gojq, sha256sum
# and install on PATH.
set -euo pipefail
shopt -s inherit_errexit

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
issuer='https://token.actions.githubusercontent.com'

fail() {
  echo "release-fetch: $*" >&2
  exit 1
}

say() {
  echo "release-fetch: $*"
}

usage='usage: CNPG_VERSION=<x.y.z> CNPG_SHA256=<sha256> release-fetch.sh <version> <owner/repo> <empty dir>'
[ "$#" -eq 3 ] || {
  echo "$usage" >&2
  exit 2
}
VERSION="$1"
REPO="$2"
dir="$3"
OS="${OS:-linux}"
ARCH="${ARCH:-amd64}"
: "${CNPG_VERSION:?$usage}" "${CNPG_SHA256:?$usage}"
[[ "$VERSION" =~ ^[0-9]{4}\.[0-9]+\.[0-9]+(-(alpha|beta)\.[0-9]+)?$ ]] || fail "'$VERSION' is not a release version"
[[ "$REPO" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail "'$REPO' is not owner/repo"
[[ "$CNPG_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "'$CNPG_VERSION' is not a CloudNativePG version"
[[ "$CNPG_SHA256" =~ ^[0-9a-f]{64}$ ]] || fail "'$CNPG_SHA256' is not a sha256"
[[ "$OS" =~ ^(linux|darwin)$ && "$ARCH" =~ ^(amd64|arm64)$ ]] || fail "$OS/$ARCH is not a platform the release ships a CLI for"
mkdir -p "$dir"
[ -z "$(ls -A "$dir")" ] || fail "$dir is not empty: the assets go into a new, empty directory"
cd "$dir"

images_id="https://github.com/$REPO/.github/workflows/images.yml@refs/heads/main"
cli="openfief-$VERSION-$OS-$ARCH"
sums="openfief-$VERSION-SHA256SUMS"
sbom="openfief-$VERSION.cdx.json"
cnpg="cnpg-$CNPG_VERSION.yaml"

# 1. Download the assets.
say "1. downloading $VERSION's assets from $REPO"
bash "$root/scripts/retry.sh" gh release download "$VERSION" --repo "$REPO" \
  --pattern "openfief-crds-$VERSION.yaml" --pattern "openfief-install-$VERSION.yaml" \
  --pattern "openfief-db-$VERSION.yaml" --pattern "$cnpg"
mkdir cli
# The signed checksums file lists the CLI of every platform, and release-verify.sh checks it strictly, so every
# platform's CLI and signature is downloaded, not only the one OS and ARCH name.
patterns=(--pattern "$sums" --pattern "$sums.sigstore.json" --pattern "$sbom" --pattern "$sbom.sigstore.json")
for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do
  patterns+=(--pattern "openfief-$VERSION-$platform" --pattern "openfief-$VERSION-$platform.sigstore.json")
done
bash "$root/scripts/retry.sh" gh release download "$VERSION" --repo "$REPO" --dir cli "${patterns[@]}"
for f in "openfief-crds-$VERSION.yaml" "openfief-install-$VERSION.yaml" "openfief-db-$VERSION.yaml" "$cnpg" \
  "cli/$cli" "cli/$cli.sigstore.json" "cli/$sums" "cli/$sums.sigstore.json" "cli/$sbom" "cli/$sbom.sigstore.json"; do
  [ -s "$f" ] || fail "the release has no asset ${f#cli/}"
done

# 2. Verify the CLI, the SBOM and the checksums; cosign must refuse a wrong identity.
say "2. verifying the CLI as release.yml on main of $REPO signed it"
bash "$root/scripts/release-verify.sh" cli cli "$REPO" || fail "the CLI files don't verify"
grep -q "  $cli\$" "cli/$sums" || fail "$sums doesn't list $cli"
if cosign verify-blob --bundle "cli/$cli.sigstore.json" --certificate-identity "$images_id" \
  --certificate-oidc-issuer "$issuer" --certificate-github-workflow-repository "$REPO" \
  --certificate-github-workflow-ref refs/heads/main "cli/$cli" >/dev/null 2>&1; then
  fail "cosign accepted $cli as signed by $images_id, a wrong identity"
fi
say "cosign refused $cli under a wrong identity"
mkdir bin
install -m 0755 "cli/$cli" bin/openfief

# 3. Verify the images the install runs.
say "3. verifying the install's images as images.yml on main signed them"
bash "$root/scripts/release-verify.sh" images "openfief-install-$VERSION.yaml" "$REPO" || fail "the install's images don't verify"

# 4. CloudNativePG's manifest is the pinned one.
say "4. checking $cnpg against the pinned sha256"
echo "$CNPG_SHA256  $cnpg" | sha256sum --check || fail "$cnpg doesn't match the pinned sha256"

# 5. The release's demo-seed image, by the digest cosign verified.
prefix="ghcr.io/${REPO,,}"
say "5. resolving $prefix/demo-seed:$VERSION to the digest images.yml on main signed"
digest="$(bash "$root/scripts/verified-digest.sh" "$prefix/demo-seed:$VERSION" "$REPO")" ||
  fail "$prefix/demo-seed:$VERSION has no signature from $REPO's images workflow on main"
echo "$prefix/demo-seed@$digest" >demo-seed.image
say "verified $prefix/demo-seed@$digest"
say "$VERSION fetched and verified into $dir"
