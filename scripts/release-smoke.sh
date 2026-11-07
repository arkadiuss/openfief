#!/usr/bin/env bash
# Installs a PUBLISHED release the way the install guide tells the owner to (docs/guides/install.md), on the cluster
# of the current kubeconfig context, and fails on the first step that doesn't do what the guide says it does. It
# downloads the release's own assets from its GitHub release, never anything this checkout builds:
#   1. downloads the assets with gh;
#   2. verifies the CLI, the checksums file and the SBOM with cosign verify-blob as release.yml on main signed them,
#      checks the checksums, and checks that cosign REFUSES the CLI under a wrong identity; puts the CLI on PATH and,
#      unless CLI_VERSION=0, checks that `openfief version` names the version and the commit the tag points at;
#   3. verifies every image the install file runs with cosign verify as images.yml on main signed it, each pinned by
#      digest, and checks that cosign REFUSES the first one under the CLI's identity;
#   4. checks CloudNativePG's manifest against the pinned sha256 and installs it;
#   5. applies the CRDs, the database and the install, and waits for the database and every rollout;
#   6. runs `openfief plugin install ui --yes` and waits for the UI plugin to be Ready;
#   7. runs `openfief enroll`, reads the one-time link it prints, has Chromium create the first passkey through it
#      and sign in again (PASSKEY, test/e2e/passkey.mjs, as the end-to-end suite does), then stops it as Ctrl-C does
#      and checks it exits cleanly;
#   8. audits the egress proxy's log against the PluginGrants with `egress-proxy audit` (EGRESS_AUDIT, built from this
#      checkout, not the release): every request it allowed went to a host the asking instance's grant approves, and
#      no line holds more than who asked, where to and the decision (D-200 criterion 16, GR-12). It reads only, and
#      prints counts, instances and hosts.
# Beyond the guide, step 2 also verifies the SBOM, step 8 is the audit, and the negative checks prove the identity
# flags are not ignored.
# On a failure it prints the cluster's state (pods, events, plugins, logs; never a Secret).
#
# Usage: CNPG_VERSION=<x.y.z> CNPG_SHA256=<sha256> PASSKEY=<passkey.mjs> EGRESS_AUDIT=<egress-proxy binary> \
#          release-smoke.sh <version> <owner/repo> <empty work dir>
# Optional: OS and ARCH of the CLI to download (default linux, amd64); CLI_VERSION=0 skips the `openfief version`
# check, for releases that predate the command (2026.11.0-alpha.1). Needs gh (GH_TOKEN), cosign, kubectl, node,
# sha256sum and install on PATH.
set -euo pipefail
shopt -s inherit_errexit

issuer='https://token.actions.githubusercontent.com'
link_timeout="${LINK_TIMEOUT:-120}"

fail() {
  echo "release-smoke: $*" >&2
  exit 1
}

say() {
  echo "release-smoke: $*"
}

usage='usage: CNPG_VERSION=<x.y.z> CNPG_SHA256=<sha256> PASSKEY=<passkey.mjs> EGRESS_AUDIT=<egress-proxy binary> release-smoke.sh <version> <owner/repo> <empty work dir>'
[ "$#" -eq 3 ] || {
  echo "$usage" >&2
  exit 2
}
VERSION="$1"
REPO="$2"
work="$3"
OS="${OS:-linux}"
ARCH="${ARCH:-amd64}"
: "${CNPG_VERSION:?$usage}" "${CNPG_SHA256:?$usage}" "${PASSKEY:?$usage}" "${EGRESS_AUDIT:?$usage}"
[[ "$VERSION" =~ ^[0-9]{4}\.[0-9]+\.[0-9]+(-(alpha|beta)\.[0-9]+)?$ ]] || fail "'$VERSION' is not a release version"
[[ "$REPO" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail "'$REPO' is not owner/repo"
[[ "$CNPG_SHA256" =~ ^[0-9a-f]{64}$ ]] || fail "'$CNPG_SHA256' is not a sha256"
PASSKEY="$(cd "$(dirname "$PASSKEY")" && pwd)/$(basename "$PASSKEY")"
[ -f "$PASSKEY" ] || fail "$PASSKEY is not a file"
EGRESS_AUDIT="$(cd "$(dirname "$EGRESS_AUDIT")" && pwd)/$(basename "$EGRESS_AUDIT")"
[ -x "$EGRESS_AUDIT" ] || fail "$EGRESS_AUDIT is not an executable"
mkdir -p "$work"
[ -z "$(ls -A "$work")" ] || fail "$work is not empty: the assets go into a new, empty directory"
cd "$work"

cli_id="https://github.com/$REPO/.github/workflows/release.yml@refs/heads/main"
images_id="https://github.com/$REPO/.github/workflows/images.yml@refs/heads/main"
cli="openfief-$VERSION-$OS-$ARCH"
sums="openfief-$VERSION-SHA256SUMS"
sbom="openfief-$VERSION.cdx.json"
cnpg="cnpg-$CNPG_VERSION.yaml"
cluster_touched=0
enroll_pid=""

# diagnose prints what an owner would look at when a step on the cluster fails. Secrets are never read.
diagnose() {
  echo "release-smoke: diagnostics" >&2
  kubectl get pods -A -o wide >&2 || true
  kubectl -n openfief get events --sort-by=.lastTimestamp >&2 || true
  kubectl -n openfief get plugins.openfief.io,plugingrants.openfief.io,clusters.postgresql.cnpg.io -o yaml >&2 || true
  kubectl -n openfief describe pods >&2 || true
  for d in openfief-operator egress-proxy openfief-core; do
    kubectl -n openfief logs "deployment/$d" --all-containers --tail=200 >&2 || true
  done
  kubectl -n cnpg-system logs deployment/cnpg-controller-manager --tail=100 >&2 || true
}

finish() {
  local status=$?
  if [ -n "$enroll_pid" ]; then
    kill -TERM "$enroll_pid" 2>/dev/null || true
    wait "$enroll_pid" 2>/dev/null || true
  fi
  if [ "$status" -ne 0 ] && [ "$cluster_touched" = 1 ]; then
    diagnose
  fi
  exit "$status"
}
trap finish EXIT

# verify_blob is the guide's cosign verify-blob for file $1 as identity $2.
verify_blob() {
  cosign verify-blob --bundle "$1.sigstore.json" \
    --certificate-identity "$2" \
    --certificate-oidc-issuer "$issuer" \
    --certificate-github-workflow-repository "$REPO" \
    --certificate-github-workflow-ref refs/heads/main \
    "$1"
}

# verify_image is the guide's cosign verify for image $1 as identity $2.
verify_image() {
  cosign verify "$1" \
    --certificate-identity "$2" \
    --certificate-oidc-issuer "$issuer" \
    --certificate-github-workflow-repository "$REPO" \
    --certificate-github-workflow-ref refs/heads/main >/dev/null
}

# 1. Download the assets.
say "1. downloading $VERSION's assets from $REPO"
gh release download "$VERSION" --repo "$REPO" \
  --pattern "openfief-crds-$VERSION.yaml" --pattern "openfief-install-$VERSION.yaml" \
  --pattern "openfief-db-$VERSION.yaml" --pattern "$cnpg" \
  --pattern "$cli" --pattern "$cli.sigstore.json" \
  --pattern "$sums" --pattern "$sums.sigstore.json" \
  --pattern "$sbom" --pattern "$sbom.sigstore.json"
for f in "openfief-crds-$VERSION.yaml" "openfief-install-$VERSION.yaml" "openfief-db-$VERSION.yaml" "$cnpg" \
  "$cli" "$cli.sigstore.json" "$sums" "$sums.sigstore.json" "$sbom" "$sbom.sigstore.json"; do
  [ -s "$f" ] || fail "the release has no asset $f"
done

# 2. Verify the CLI.
say "2. verifying the CLI as $cli_id"
for f in "$cli" "$sums" "$sbom"; do
  verify_blob "$f" "$cli_id" || fail "$f: its signature doesn't verify as $cli_id"
done
sha256sum --check --ignore-missing "$sums" || fail "a checksum in $sums doesn't match"
grep -q "  $cli\$" "$sums" || fail "$sums doesn't list $cli"
if verify_blob "$cli" "$images_id" >/dev/null 2>&1; then
  fail "cosign accepted $cli as signed by $images_id, a wrong identity"
fi
say "cosign refused $cli under a wrong identity"
mkdir -p bin
install -m 0755 "$cli" bin/openfief
PATH="$PWD/bin:$PATH"
if [ "${CLI_VERSION:-1}" = 1 ]; then
  commit="$(gh api "repos/$REPO/commits/$VERSION" --jq .sha)" || fail "can't read the commit $VERSION tags"
  [[ "$commit" =~ ^[0-9a-f]{40}$ ]] || fail "'$commit' is not the commit $VERSION tags"
  got="$(openfief version)" || fail "openfief version failed"
  [ "$got" = "openfief $VERSION commit $commit" ] || fail "openfief version printed '$got', want 'openfief $VERSION commit $commit'"
  say "$got"
fi

# 3. Verify the images.
say "3. verifying the install's images as $images_id"
install_file="openfief-install-$VERSION.yaml"
mapfile -t images < <(sed -n 's/^ *image: *//p' "$install_file" | sort -u)
[ "${#images[@]}" -gt 0 ] || fail "$install_file runs no image"
[ "$(grep -c 'image:' "$install_file")" -eq "$(sed -n 's/^ *image: *//p' "$install_file" | wc -l)" ] ||
  fail "$install_file names an image the guide's command doesn't read"
for image in "${images[@]}"; do
  [[ "$image" =~ ^[a-z0-9][a-z0-9._/-]*@sha256:[0-9a-f]{64}$ ]] || fail "$install_file runs $image, not an image by digest"
  verify_image "$image" "$images_id" || fail "$image: its signature doesn't verify as $images_id"
  say "verified $image"
done
if verify_image "${images[0]}" "$cli_id" 2>/dev/null; then
  fail "cosign accepted ${images[0]} as signed by $cli_id, a wrong identity"
fi
say "cosign refused ${images[0]} under a wrong identity"

# 4. Install CloudNativePG.
say "4. installing CloudNativePG $CNPG_VERSION"
echo "$CNPG_SHA256  $cnpg" | sha256sum --check || fail "$cnpg doesn't match the pinned sha256"
cluster_touched=1
kubectl apply --server-side -f "$cnpg"
kubectl -n cnpg-system rollout status deployment/cnpg-controller-manager --timeout=5m

# 5. Apply the CRDs, the database and the install.
say "5. applying the CRDs, the database and the install"
kubectl apply --server-side -f "openfief-crds-$VERSION.yaml"
kubectl apply --server-side -f "openfief-db-$VERSION.yaml"
kubectl -n openfief wait cluster.postgresql.cnpg.io/openfief-db --for=condition=Ready --timeout=10m
kubectl apply --server-side -f "$install_file"
for d in openfief-operator egress-proxy openfief-core; do
  kubectl -n openfief rollout status "deployment/$d" --timeout=5m
done

# 6. Install the built-in UI.
say "6. openfief plugin install ui"
openfief plugin install ui --yes
kubectl -n openfief wait plugin.openfief.io/ui --for=condition=Ready --timeout=5m

# 7. Create the first passkey.
say "7. openfief enroll and the first passkey"
openfief enroll >enroll.out 2>enroll.err &
enroll_pid=$!
link=""
for _ in $(seq "$link_timeout"); do
  link="$(grep -m1 -o 'http://localhost:8080/enroll#code=[^[:space:]]*' enroll.out || true)"
  [ -n "$link" ] && break
  kill -0 "$enroll_pid" 2>/dev/null || fail "openfief enroll exited before printing a link: $(cat enroll.err)"
  sleep 1
done
[ -n "$link" ] || fail "openfief enroll printed no link in ${link_timeout}s: $(cat enroll.err)"
say "openfief enroll printed the one-time link"
OPENFIEF_E2E_ENROLL_LINK="$link" node "$PASSKEY" || fail "the browser couldn't create the first passkey and sign in with it"
# Ctrl-C's signal is ignored by a background job of a script, so stop it with SIGTERM, which it handles the same way.
kill -TERM "$enroll_pid"
status=0
wait "$enroll_pid" || status=$?
enroll_pid=""
[ "$status" -eq 0 ] || fail "openfief enroll exited with $status when stopped: $(cat enroll.err)"

# 8. Audit the egress proxy's log.
say "8. auditing the egress proxy's log against the PluginGrants"
"$EGRESS_AUDIT" audit || fail "the egress proxy's log breaks D-200's criterion 16 or GR-12 (see the audit above)"
say "$VERSION installs as the guide says"
