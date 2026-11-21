#!/usr/bin/env bash
# Tests for scripts/release-fetch.sh with stand-ins for gh and cosign and a fake release (kibi:
# TEST-tooling-release-fetch): a release whose signatures verify is downloaded into an empty directory with its CLI
# executable and its demo-seed image recorded by digest; a wrong identity accepted, a bad checksum, a tampered SBOM, a
# missing asset, an unsigned or unpinned image, an unsigned demo-seed image, a CloudNativePG manifest that isn't the
# pinned one and a directory that isn't empty each fail it. Offline: no registry, Sigstore or GitHub call.
# Needs gojq on PATH (task test adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
repo=example/openfief
version=2026.11.0-alpha.1
cli_id="https://github.com/$repo/.github/workflows/release.yml@refs/heads/main"
images_id="https://github.com/$repo/.github/workflows/images.yml@refs/heads/main"
op="ghcr.io/$repo/operator@sha256:$(printf '1%.0s' {1..64})"
eg="ghcr.io/$repo/egress-proxy@sha256:$(printf '2%.0s' {1..64})"
co="ghcr.io/$repo/core@sha256:$(printf '3%.0s' {1..64})"
seed_tag="ghcr.io/$repo/demo-seed:$version"
seed_digest="sha256:$(printf '7%.0s' {1..64})"
mkdir -p "$work/bin"

# release makes the fake release's assets in $work/release: the bundle, with $1 as the install's images, and the
# CLI, checksums and SBOM, each with a bundle the fake cosign accepts for release.yml on main.
release() {
  local r="$work/release" f
  rm -rf "$r"
  mkdir -p "$r"
  printf 'kind: Namespace\nmetadata:\n  name: openfief\n' >"$r/openfief-crds-$version.yaml"
  printf 'kind: Cluster\nmetadata:\n  name: openfief-db\n' >"$r/openfief-db-$version.yaml"
  printf 'kind: Namespace\nmetadata:\n  name: cnpg-system\n' >"$r/cnpg-1.30.1.yaml"
  : >"$r/openfief-install-$version.yaml"
  for f in $1; do
    printf -- '---\nkind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - name: c\n          image: %s\n' "$f" >>"$r/openfief-install-$version.yaml"
  done
  printf '#!/usr/bin/env bash\necho the release CLI\n' >"$r/openfief-$version-linux-amd64"
  echo '{"bomFormat":"CycloneDX"}' >"$r/openfief-$version.cdx.json"
  (cd "$r" && sha256sum "openfief-$version-linux-amd64" "openfief-$version.cdx.json" >"openfief-$version-SHA256SUMS")
  for f in "openfief-$version-linux-amd64" "openfief-$version.cdx.json" "openfief-$version-SHA256SUMS"; do
    echo "$cli_id $repo refs/heads/main $(sha256sum <"$r/$f" | cut -d' ' -f1)" >"$r/$f.sigstore.json"
  done
}

# gh copies each --pattern asset from $work/release into --dir (default the current directory).
cat >"$work/bin/gh" <<'EOF'
#!/usr/bin/env bash
echo "gh $*" >>"$FAKE_LOG"
[ "$1 $2" = "release download" ] || exit 2
shift 2
dest=.
while [ $# -gt 0 ]; do
  case "$1" in
    --dir) dest="$2"; shift 2 ;;
    --pattern) [ ! -f "$FAKE_RELEASE/$2" ] || cp "$FAKE_RELEASE/$2" "$dest"; shift 2 ;;
    *) shift ;;
  esac
done
EOF
# cosign verify-blob passes when the bundle holds "<identity> <repository> <ref> <sha256 of the file>"; verify
# "verifies" the refs in FAKE_SIGNED (ref=digest) as images.yml on main and prints the digest as cosign's JSON does.
# FAKE_LAX=1 accepts any identity.
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "cosign $*" >>"$FAKE_LOG"
cmd="$1"
shift
bundle="" id="" iss="" repo="" ref="" target=""
while [ $# -gt 0 ]; do
  case "$1" in
    --bundle) bundle="$2"; shift 2 ;;
    --certificate-identity) id="$2"; shift 2 ;;
    --certificate-oidc-issuer) iss="$2"; shift 2 ;;
    --certificate-github-workflow-repository) repo="$2"; shift 2 ;;
    --certificate-github-workflow-ref) ref="$2"; shift 2 ;;
    --output) shift 2 ;;
    *) target="$1"; shift ;;
  esac
done
[ "$iss" = https://token.actions.githubusercontent.com ] && [ -n "$repo" ] && [ -n "$ref" ] || exit 1
case "$cmd" in
  verify-blob)
    want="$(cat "$bundle")"
    [ "${FAKE_LAX:-0}" = 1 ] && id="${want%% *}"
    [ "$want" = "$id $repo $ref $(sha256sum <"$target" | cut -d' ' -f1)" ]
    ;;
  verify)
    [ "${FAKE_LAX:-0}" = 1 ] || [ "$id" = "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main" ] || exit 1
    for entry in $FAKE_SIGNED; do
      if [ "${entry%%=*}" = "$target" ]; then
        printf '[{"critical":{"image":{"docker-manifest-digest":"%s"}}}]\n' "${entry#*=}"
        exit 0
      fi
    done
    exit 1
    ;;
  *) exit 2 ;;
esac
EOF
chmod +x "$work/bin/"*
cnpg_sum() { sha256sum <"$work/release/cnpg-1.30.1.yaml" | cut -d' ' -f1; }
signed() { echo "${op}=${op#*@} ${eg}=${eg#*@} ${co}=${co#*@} $seed_tag=$seed_digest"; }

# run runs release-fetch.sh into a new directory with the fakes; its output lands in $work/out, the calls in $work/log.
# Extra environment (VAR=value) comes first.
run() {
  : >"$work/log"
  rm -rf "$work/d"
  env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_RELEASE="$work/release" FAKE_SIGNED="$(signed)" \
    CNPG_VERSION=1.30.1 CNPG_SHA256="$(cnpg_sum)" \
    "$@" bash "$root/scripts/release-fetch.sh" "$version" "$repo" "$work/d" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the last run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}
yes_if() { if "$@"; then echo yes; else echo no; fi; }
passes() { if run "$@"; then echo yes; else echo no; fi; }
# fails_with runs with the environment after the pattern and wants a failure whose output matches the pattern.
fails_with() {
  local pattern="$1"
  shift
  if run "$@"; then echo no; elif grep -q -- "$pattern" "$work/out"; then echo yes; else echo no; fi
}

release "$op $eg $co"
check "a release that verifies is fetched" "$(passes)"
for f in "openfief-crds-$version.yaml" "openfief-install-$version.yaml" "openfief-db-$version.yaml" cnpg-1.30.1.yaml; do
  check "the bundle's $f is in the directory" "$(yes_if test -s "$work/d/$f")"
done
check "the verified CLI is executable under bin/" "$(yes_if test -x "$work/d/bin/openfief")"
check "the CLI is the downloaded file" "$(yes_if cmp -s "$work/d/bin/openfief" "$work/release/openfief-$version-linux-amd64")"
check "the demo-seed image is recorded by the verified digest" "$(yes_if test "$(cat "$work/d/demo-seed.image")" = "ghcr.io/$repo/demo-seed@$seed_digest")"
check "it verifies the CLI, the checksums and the SBOM as release.yml on main" "$(yes_if test "$(grep -c "^cosign verify-blob .*--certificate-identity $cli_id " "$work/log")" -eq 3)"
check "it pins the run's repository and ref on every check" "$(yes_if test "$(grep -c -- "--certificate-github-workflow-repository $repo --certificate-github-workflow-ref refs/heads/main" "$work/log")" -ge 7)"
check "it tries the CLI under a wrong identity" "$(yes_if grep -q "^cosign verify-blob .*--certificate-identity $images_id .*openfief-$version-linux-amd64\$" "$work/log")"
check "it verifies each image of ours as images.yml" "$(yes_if test "$(grep -c "^cosign verify ghcr.io/$repo/.*@sha256:.* --certificate-identity $images_id " "$work/log")" -eq 3)"
check "it verifies the demo-seed image by its tag as images.yml" "$(yes_if grep -q "^cosign verify $seed_tag .*--certificate-identity $images_id " "$work/log")"
check "the CLI and the bundle come from the release, by gh" "$(yes_if grep -q "^gh release download $version --repo $repo --dir cli " "$work/log")"
check "OS and ARCH pick the CLI file" "$(yes_if grep -q -- "--pattern openfief-$version-linux-amd64 " "$work/log")"

mkdir -p "$work/full"
touch "$work/full/x"
if PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_RELEASE="$work/release" FAKE_SIGNED="$(signed)" CNPG_VERSION=1.30.1 CNPG_SHA256="$(cnpg_sum)" \
  bash "$root/scripts/release-fetch.sh" "$version" "$repo" "$work/full" >"$work/out" 2>&1; then
  check "a directory that isn't empty is refused" no
else
  check "a directory that isn't empty is refused" "$(yes_if grep -q 'is not empty' "$work/out")"
fi
if PATH="$work/bin:$PATH" CNPG_VERSION=1.30.1 CNPG_SHA256="$(cnpg_sum)" \
  bash "$root/scripts/release-fetch.sh" "main" "$repo" "$work/new" >"$work/out" 2>&1; then
  check "a version that isn't a release is refused" no
else
  check "a version that isn't a release is refused" "$(yes_if grep -q 'is not a release version' "$work/out")"
fi

check "a wrong identity accepted fails" "$(fails_with 'a wrong identity' FAKE_LAX=1)"
check "a CloudNativePG manifest that isn't pinned fails" "$(fails_with "doesn't match the pinned sha256" CNPG_SHA256="$(printf 'a%.0s' {1..64})")"
check "an unsigned demo-seed image fails" "$(fails_with "has no signature" FAKE_SIGNED="$op=${op#*@} $eg=${eg#*@} $co=${co#*@}")"
check "an unsigned image of the install fails" "$(fails_with "install's images don't verify" FAKE_SIGNED="$op=${op#*@} $eg=${eg#*@} $seed_tag=$seed_digest")"

release "$op ghcr.io/$repo/core:main"
check "an image by tag fails" "$(fails_with "install's images don't verify")"
release "$op $eg $co"
echo tampered >>"$work/release/openfief-$version.cdx.json"
check "a tampered SBOM fails" "$(fails_with "CLI files don't verify")"
release "$op $eg $co"
echo tampered >>"$work/release/openfief-$version-linux-amd64"
check "a tampered CLI fails" "$(fails_with "CLI files don't verify")"
release "$op $eg $co"
rm "$work/release/openfief-$version-linux-amd64.sigstore.json"
check "a missing asset fails" "$(fails_with "the release has no asset openfief-$version-linux-amd64.sigstore.json")"
release "$op $eg $co"
rm "$work/release/openfief-db-$version.yaml"
check "a release without its database fails" "$(fails_with "the release has no asset openfief-db-$version.yaml")"
third="quay.io/authzed/spicedb:v1.57.1@sha256:$(printf 'a%.0s' {1..64})"
release "$op $eg $co $third"
check "a third party's image pinned by tag and digest passes" "$(passes)"
release "$op $eg $co quay.io/authzed/spicedb:v1.57.1"
check "a third party's image without its digest fails" "$(fails_with "install's images don't verify")"

if [ "$failures" -ne 0 ]; then
  echo "release-fetch: $failures of $cases checks failed" >&2
  exit 1
fi
echo "release-fetch: $cases checks passed"
