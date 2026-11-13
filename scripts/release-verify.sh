#!/usr/bin/env bash
# Verifies a release's signatures exactly as the install guide tells the owner to (docs/guides/install.md), with
# cosign and the keyless identities the release signs with:
#   - the CLI files are signed in release.yml's cli job (task release:cli:sign), so their certificate names
#     release.yml at the ref the run started from, refs/heads/main for a published release;
#   - the images are signed in images.yml's build job (task image:sign), a reusable workflow release.yml calls, and
#     Fulcio names the reusable workflow, so their certificate names images.yml at refs/heads/main.
# Both come from GitHub Actions' OIDC issuer. Every check also pins the repository and ref of the run that signed
# (cosign's --certificate-github-workflow-repository and --certificate-github-workflow-ref, D-181): images.yml is a
# reusable workflow in a public repository, so another repository's workflow can call it and get a certificate
# with the very same identity; only the run's repository and ref tell OpenFief's own run apart.
#
# Usage:
#   release-verify.sh identity <cli|images> <owner/repo> [ref]
#       prints the certificate identity; ref (cli only) defaults to refs/heads/main
#   release-verify.sh cli <dir> <owner/repo> [ref]
#       checks <dir>'s SHA256SUMS, then that every other file in <dir> verifies against its <file>.sigstore.json
#       bundle as release.yml at ref (default refs/heads/main) signed it
#   release-verify.sh images <install file> <owner/repo> [tag]
#       checks that each image the install runs is pinned by digest and verifies as images.yml on main signed that
#       digest; with tag (a dry run, whose install is pinned to placeholders), it verifies <repository>:<tag> of each
#       image instead, such as the images main's last build signed
# Needs cosign and gojq on PATH (cli needs only cosign and sha256sum).
set -euo pipefail
shopt -s inherit_errexit

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
issuer='https://token.actions.githubusercontent.com'

fail() {
  echo "release-verify: $*" >&2
  exit 1
}

# identity prints the certificate identity of workflow $1 (cli or images) of repository $2 at ref $3.
identity() {
  local kind="$1" repo="$2" ref="${3:-refs/heads/main}"
  [[ "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail "'$repo' is not owner/repo"
  [[ "$ref" =~ ^refs/(heads|tags)/[A-Za-z0-9_./-]+$ ]] || fail "'$ref' is not a branch or tag ref"
  case "$kind" in
    cli) echo "https://github.com/$repo/.github/workflows/release.yml@$ref" ;;
    images)
      [ "$ref" = refs/heads/main ] || fail "images are signed only from main"
      echo "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main"
      ;;
    *) fail "unknown kind '$kind' (cli or images)" ;;
  esac
}

# verify_cli checks the checksums and the signature of every file in directory $1 as release.yml of $2 at $3.
verify_cli() {
  local dir="$1" id file name count=0
  id="$(identity cli "$2" "${3:-}")"
  [ -d "$dir" ] || fail "$dir is not a directory"
  local sums=("$dir"/openfief-*-SHA256SUMS)
  [ -f "${sums[0]}" ] || fail "$dir has no openfief-<version>-SHA256SUMS"
  [ "${#sums[@]}" -eq 1 ] || fail "$dir has more than one SHA256SUMS"
  (cd "$dir" && sha256sum --check --strict --quiet "$(basename "${sums[0]}")") || fail "$dir: a checksum doesn't match"
  for file in "$dir"/*; do
    name="$(basename "$file")"
    case "$name" in *.sigstore.json) continue ;; esac
    [ -f "$file.sigstore.json" ] || fail "$name has no signature bundle $name.sigstore.json"
    cosign verify-blob --bundle "$file.sigstore.json" \
      --certificate-identity "$id" --certificate-oidc-issuer "$issuer" \
      --certificate-github-workflow-repository "$2" --certificate-github-workflow-ref "${3:-refs/heads/main}" \
      "$file" >/dev/null 2>&1 ||
      fail "$name: its signature doesn't verify as $id"
    echo "release-verify: $name signed by $id" >&2
    count=$((count + 1))
  done
  [ "$count" -gt 1 ] || fail "$dir holds no CLI files besides SHA256SUMS"
}

# verify_images checks every image install file $1 runs as images.yml on main of $2 signed it, at tag $3 if given.
verify_images() {
  local file="$1" repo="$2" tag="${3:-}" image ref digest count=0
  identity images "$repo" >/dev/null
  [ -f "$file" ] || fail "$file is not a file"
  [ -z "$tag" ] || [[ "$tag" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$ ]] || fail "'$tag' is not an image tag"
  while IFS= read -r image; do
    # A third party's image (SpiceDB and the Job that generates its credentials, D-164) is named by tag and digest and
    # signed by nobody of ours: it must be pinned, and is left out of the signature check.
    if [[ "$image" != "ghcr.io/$repo/"* && "$image" != "ghcr.io/${repo,,}/"* ]]; then
      [[ "$image" =~ ^[a-z0-9][a-z0-9._/-]*:[A-Za-z0-9_][A-Za-z0-9_.-]*@sha256:[0-9a-f]{64}$ ]] ||
        fail "the install runs $image, a third party's image that isn't pinned by tag and digest"
      echo "release-verify: $image is a third party's, pinned" >&2
      continue
    fi
    [[ "$image" =~ ^[a-z0-9][a-z0-9._/-]*@sha256:[0-9a-f]{64}$ ]] || fail "the install runs $image, not an image by digest"
    if [ -n "$tag" ]; then ref="${image%@*}:$tag"; else ref="$image"; fi
    digest="$(bash "$root/scripts/verified-digest.sh" "$ref" "$repo")" ||
      fail "$ref has no signature from $repo's images workflow on main"
    [ -n "$tag" ] || [ "$digest" = "${image#*@}" ] || fail "$image: the signature names $digest"
    echo "release-verify: $ref ($digest) signed by $(identity images "$repo")" >&2
    count=$((count + 1))
  done < <(gojq -r --yaml-input -s \
    '[.[] | .. | objects | (.containers?, .initContainers?) // empty | .[] | .image] | unique | .[]' "$file")
  [ "$count" -gt 0 ] || fail "$file runs no image"
}

usage='usage: release-verify.sh identity <cli|images> <owner/repo> [ref] | cli <dir> <owner/repo> [ref] | images <install file> <owner/repo> [tag]'
[ "$#" -ge 3 ] && [ "$#" -le 4 ] || {
  echo "$usage" >&2
  exit 2
}
command="$1"
shift
case "$command" in
  identity) identity "$@" ;;
  cli) verify_cli "$@" ;;
  images) verify_images "$@" ;;
  *)
    echo "$usage" >&2
    exit 2
    ;;
esac
