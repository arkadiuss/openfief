#!/usr/bin/env bash
# Tests for scripts/release-verify.sh with a fake cosign, and that the install guide's verification commands name
# the identities the release really signs with (kibi: TEST-tooling-release-verify). No registry or Sigstore call.
# Needs gojq on PATH (task test adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
repo=example/openfief
issuer='https://token.actions.githubusercontent.com'
cli_main="https://github.com/$repo/.github/workflows/release.yml@refs/heads/main"
images_main="https://github.com/$repo/.github/workflows/images.yml@refs/heads/main"
zero="sha256:$(printf '0%.0s' {1..64})"
op='sha256:1111111111111111111111111111111111111111111111111111111111111111'
co='sha256:3333333333333333333333333333333333333333333333333333333333333333'

# The fake cosign logs its arguments. verify-blob passes when the bundle holds
# "<identity> <source repository> <source ref> <sha256 of the file>" for the identity, workflow repository and ref
# and file it was given, with GitHub's issuer; verify "verifies" only the refs in FAKE_SIGNED (ref=digest), and only
# when it pins the run's repository and ref, and prints the digest as cosign's JSON does.
mkdir -p "$work/bin"
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$FAKE_LOG"
case "$1" in
  verify-blob)
    shift
    bundle="" id="" iss="" repo="" ref="" file=""
    while [ $# -gt 0 ]; do
      case "$1" in
        --bundle) bundle="$2"; shift 2 ;;
        --certificate-identity) id="$2"; shift 2 ;;
        --certificate-oidc-issuer) iss="$2"; shift 2 ;;
        --certificate-github-workflow-repository) repo="$2"; shift 2 ;;
        --certificate-github-workflow-ref) ref="$2"; shift 2 ;;
        *) file="$1"; shift ;;
      esac
    done
    [ "$iss" = https://token.actions.githubusercontent.com ] || exit 1
    [ -n "$repo" ] && [ -n "$ref" ] || exit 1
    [ "$(cat "$bundle")" = "$id $repo $ref $(sha256sum <"$file" | cut -d' ' -f1)" ]
    ;;
  verify)
    case "$*" in *" --certificate-github-workflow-repository "*" --certificate-github-workflow-ref refs/heads/main "*) ;; *) exit 1 ;; esac
    for entry in $FAKE_SIGNED; do
      if [ "${entry%%=*}" = "$2" ]; then
        printf '[{"critical":{"image":{"docker-manifest-digest":"%s"}}}]\n' "${entry#*=}"
        exit 0
      fi
    done
    exit 1
    ;;
  *) exit 2 ;;
esac
EOF
chmod +x "$work/bin/cosign"

# run runs release-verify.sh with the fake cosign and FAKE_SIGNED=$1, keeping its output in $work/out.
run() {
  local signed="$1"
  shift
  : >"$work/log"
  PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_SIGNED="$signed" \
    bash "$root/scripts/release-verify.sh" "$@" >"$work/out" 2>&1
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

# passes and fails print yes when a run of release-verify.sh with FAKE_SIGNED=$1 succeeds or fails.
passes() {
  if run "$@"; then echo yes; else echo no; fi
}
fails() {
  if run "$@"; then echo no; else echo yes; fi
}

# yes_if prints yes when the command succeeds.
yes_if() {
  if "$@" >/dev/null 2>&1; then echo yes; else echo no; fi
}

# cli_dir writes a release's CLI files into $work/cli, each signed as identity $1 in a run of repository $2 (default
# $repo) at ref $3 (default refs/heads/main).
cli_dir() {
  local id="$1" source="${2:-$repo}" ref="${3:-refs/heads/main}" f
  rm -rf "$work/cli"
  mkdir -p "$work/cli"
  for p in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do echo "binary $p" >"$work/cli/openfief-2026.11.0-alpha.1-$p"; done
  echo '{"bomFormat":"CycloneDX"}' >"$work/cli/openfief-2026.11.0-alpha.1.cdx.json"
  (cd "$work/cli" && sha256sum openfief-2026.11.0-alpha.1[-.]* >openfief-2026.11.0-alpha.1-SHA256SUMS)
  for f in "$work/cli"/openfief-*; do
    echo "$id $source $ref $(sha256sum <"$f" | cut -d' ' -f1)" >"$f.sigstore.json"
  done
}

# Identities.
check "the CLI identity is release.yml on main" "$(yes_if test "$(bash "$root/scripts/release-verify.sh" identity cli "$repo")" = "$cli_main")"
check "a dry run's CLI identity names its own ref" "$(yes_if test "$(bash "$root/scripts/release-verify.sh" identity cli "$repo" refs/heads/try/alpha)" = "https://github.com/$repo/.github/workflows/release.yml@refs/heads/try/alpha")"
check "the images identity is images.yml on main" "$(yes_if test "$(bash "$root/scripts/release-verify.sh" identity images "$repo")" = "$images_main")"
check "images are never trusted from another ref" "$(fails "" identity images "$repo" refs/heads/try)"
check "a malformed repository is refused" "$(fails "" identity cli 'example/openfief/../x')"
check "an unknown kind is refused" "$(fails "" identity docs "$repo")"

# The CLI.
cli_dir "$cli_main"
check "a CLI signed by release.yml on main verifies" "$(passes "" cli "$work/cli" "$repo")"
check "every file was verified with GitHub's issuer" "$(yes_if test "$(grep -c -- "--certificate-identity $cli_main --certificate-oidc-issuer $issuer" "$work/log")" -eq 6)"
check "every file was verified as a run of the repository on main" "$(yes_if test "$(grep -c -- "--certificate-github-workflow-repository $repo --certificate-github-workflow-ref refs/heads/main" "$work/log")" -eq 6)"
cli_dir "$cli_main" example/attacker
check "a CLI signed by release.yml's identity from another repository's run fails" "$(fails "" cli "$work/cli" "$repo")"
cli_dir "$cli_main" "$repo" refs/heads/feature
check "a CLI signed by release.yml's identity from another ref's run fails" "$(fails "" cli "$work/cli" "$repo")"
cli_dir "https://github.com/$repo/.github/workflows/release.yml@refs/heads/try/alpha" "$repo" refs/heads/try/alpha
check "a dry run's CLI verifies at its own ref" "$(passes "" cli "$work/cli" "$repo" refs/heads/try/alpha)"
check "a dry run's CLI doesn't verify as main" "$(fails "" cli "$work/cli" "$repo")"
cli_dir "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main"
check "a CLI signed by another workflow fails" "$(fails "" cli "$work/cli" "$repo")"
cli_dir "$cli_main"
echo tampered >>"$work/cli/openfief-2026.11.0-alpha.1-linux-amd64"
check "a changed binary fails its checksum" "$(fails "" cli "$work/cli" "$repo")"
cli_dir "$cli_main"
rm "$work/cli/openfief-2026.11.0-alpha.1-darwin-arm64.sigstore.json"
check "a file without a signature bundle fails" "$(fails "" cli "$work/cli" "$repo")"
check "the failure names the file" "$(yes_if grep -q 'openfief-2026.11.0-alpha.1-darwin-arm64 has no signature bundle' "$work/out")"
cli_dir "$cli_main"
rm "$work/cli/openfief-2026.11.0-alpha.1-SHA256SUMS"
check "a CLI without SHA256SUMS fails" "$(fails "" cli "$work/cli" "$repo")"
check "a missing directory fails" "$(fails "" cli "$work/none" "$repo")"

# The images.
cat >"$work/install.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata: {name: openfief-operator}
spec:
  template:
    spec:
      containers:
        - name: operator
          image: ghcr.io/example/openfief/operator@$op
---
apiVersion: apps/v1
kind: Deployment
metadata: {name: openfief-core}
spec:
  template:
    spec:
      initContainers:
        - name: migrate
          image: ghcr.io/example/openfief/core@$co
      containers:
        - name: core
          image: ghcr.io/example/openfief/core@$co
EOF
signed="ghcr.io/example/openfief/operator@$op=$op ghcr.io/example/openfief/core@$co=$co"
check "an install of signed digests verifies" "$(passes "$signed" images "$work/install.yaml" "$repo")"
check "each image is verified once, as images.yml on main" "$(yes_if test "$(grep -c -- "--certificate-identity $images_main --certificate-oidc-issuer $issuer" "$work/log")" -eq 2)"
check "each image is verified as a run of the repository on main" "$(yes_if test "$(grep -c -- "--certificate-github-workflow-repository $repo --certificate-github-workflow-ref refs/heads/main" "$work/log")" -eq 2)"
check "an unsigned digest fails" "$(fails "ghcr.io/example/openfief/operator@$op=$op" images "$work/install.yaml" "$repo")"
check "the failure names the image" "$(yes_if grep -q "ghcr.io/example/openfief/core@$co has no signature" "$work/out")"
check "a signature for another digest fails" "$(fails "ghcr.io/example/openfief/operator@$op=$co ghcr.io/example/openfief/core@$co=$co" images "$work/install.yaml" "$repo")"
sed "s/@sha256:[0-9a-f]*/@$zero/" "$work/install.yaml" >"$work/dry.yaml"
check "a dry run's placeholders verify at the given tag" "$(passes "ghcr.io/example/openfief/operator:main=$op ghcr.io/example/openfief/core:main=$co" images "$work/dry.yaml" "$repo" main)"
check "the tag is what gets verified" "$(yes_if grep -q '^verify ghcr.io/example/openfief/core:main ' "$work/log")"
check "a dry run's placeholders don't verify by digest" "$(fails "ghcr.io/example/openfief/operator:main=$op ghcr.io/example/openfief/core:main=$co" images "$work/dry.yaml" "$repo")"
check "a tag nothing signed fails" "$(fails "ghcr.io/example/openfief/operator:main=$op" images "$work/dry.yaml" "$repo" main)"
sed "s/@$op/:2026.11.0-alpha.1/" "$work/install.yaml" >"$work/tagged.yaml"
check "an image by tag in the install fails" "$(fails "$signed" images "$work/tagged.yaml" "$repo")"
check "a malformed tag is refused" "$(fails "" images "$work/dry.yaml" "$repo" 'main;x')"
# A third party's image is checked to be pinned by tag and digest, and left out of the signature check.
{
  cat "$work/install.yaml"
  printf -- '---\napiVersion: apps/v1\nkind: Deployment\nmetadata: {name: openfief-spicedb}\nspec:\n  template:\n    spec:\n      containers:\n        - name: spicedb\n          image: quay.io/authzed/spicedb:v1.57.1@%s\n' "$op"
} >"$work/third.yaml"
check "an install with a third party's image pinned by tag and digest verifies" "$(passes "$signed" images "$work/third.yaml" "$repo")"
check "the third party's image is never asked of cosign" "$(yes_if test "$(grep -c 'quay.io' "$work/log")" -eq 0)"
sed "s|spicedb:v1.57.1@$op|spicedb@$op|" "$work/third.yaml" >"$work/third-notag.yaml"
check "a third party's image without its tag fails" "$(fails "$signed" images "$work/third-notag.yaml" "$repo")"
sed "s|spicedb:v1.57.1@$op|spicedb:v1.57.1|" "$work/third.yaml" >"$work/third-nodigest.yaml"
check "a third party's image without its digest fails" "$(fails "$signed" images "$work/third-nodigest.yaml" "$repo")"
printf 'apiVersion: v1\nkind: Namespace\nmetadata: {name: openfief}\n' >"$work/empty.yaml"
check "an install that runs no image fails" "$(fails "" images "$work/empty.yaml" "$repo")"
check "bad arguments are refused" "$(fails "" images "$work/install.yaml")"

# The install guide names exactly these identities, for this repository, and they are where the release signs.
guide="$root/docs/guides/install.md"
own_cli="$(bash "$root/scripts/release-verify.sh" identity cli arkadiuss/openfief)"
own_images="$(bash "$root/scripts/release-verify.sh" identity images arkadiuss/openfief)"
check "the guide verifies the CLI as release.yml on main" "$(yes_if grep -qF -- "--certificate-identity $own_cli" "$guide")"
check "the guide verifies the images as images.yml on main" "$(yes_if grep -qF -- "--certificate-identity $own_images" "$guide")"
check "the guide names no other identity" "$(yes_if test -z "$(grep -o -- '--certificate-identity [^ ]*' "$guide" | grep -vxF -e "--certificate-identity $own_cli" -e "--certificate-identity $own_images")")"
check "the guide names GitHub's issuer and no other" "$(yes_if test "$(grep -o -- '--certificate-oidc-issuer [^ ]*' "$guide" | sort -u)" = "--certificate-oidc-issuer $issuer")"
check "the guide pins every check to this repository's run" "$(yes_if test "$(grep -c -- '--certificate-github-workflow-repository arkadiuss/openfief \\$' "$guide")" -eq 2)"
check "the guide pins every check to main's run" "$(yes_if test "$(grep -c -- '--certificate-github-workflow-ref refs/heads/main ' "$guide")" -eq 2)"
check "the guide pins no other repository or ref" "$(yes_if test "$(grep -o -- '--certificate-github-workflow-[a-z]* [^ ]*' "$guide" | sort -u | tr '\n' ' ')" = "--certificate-github-workflow-ref refs/heads/main --certificate-github-workflow-repository arkadiuss/openfief ")"
check "the operator pins the same source repository and ref" "$(yes_if grep -qF 'OpenFiefCISourceRepository = "https://github.com/arkadiuss/openfief"' "$root/internal/plugins/trust.go")"
check "the operator trusts the same images identity" "$(yes_if grep -qF "OpenFiefCISubject = \"$own_images\"" "$root/internal/plugins/trust.go")"
check "release.yml signs the CLI" "$(yes_if grep -q 'run: task release:cli:sign' "$root/.github/workflows/release.yml")"
check "images.yml signs the images" "$(yes_if grep -q 'run: task image:sign' "$root/.github/workflows/images.yml")"
check "release.yml publishes only from main" "$(yes_if grep -qF "github.ref != 'refs/heads/main'" "$root/.github/workflows/release.yml")"

if [ "$failures" -ne 0 ]; then
  echo "release-verify: $failures of $cases checks failed" >&2
  exit 1
fi
echo "release-verify: $cases checks passed"
