#!/usr/bin/env bash
# Decides which images a CI run builds (docs/dev/pipeline.md#images). It writes what tools/affected needs to a
# temporary directory and prints the tool's two GitHub outputs, images= (the matrix to build) and carry= (the
# unchanged images whose signed digest is tagged for this commit); the tool says why on stderr.
#   reuse <owner/repo>  a push to main: for each image, the newest of HEAD's last <depth> first-parent commits
#                       whose sha-<commit> tag has a signature from the repository's images workflow on main is
#                       its earlier build; the image is carried if nothing that goes into it changed since.
#   base <revision>     a pull request: build the images that <revision>..HEAD changes, skip the rest.
#   all -               build every image.
# Usage: ci-images-select.sh <registry prefix> <reuse|base|all> <owner/repo|revision|-> "<Go module dirs>" [depth]
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"

usage() {
  echo "ci-images-select: usage: ci-images-select.sh <prefix> <reuse <owner/repo>|base <revision>|all -> \"<modules>\" [depth]" >&2
  exit 1
}

prefix="${1:-}"
mode="${2:-}"
arg="${3:-}"
modules="${4:-.}"
depth="${5:-20}"
[[ "$depth" =~ ^[1-9][0-9]*$ ]] || usage
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/diff"
: >"$work/builds"

bash "$here/ci-images.sh" "$prefix" all >"$work/matrix.json"
read -r -a module_dirs <<<"$modules"
bash "$here/go-packages.sh" "$work" "${module_dirs[@]}"

# record_build notes an image's earlier build and the changes since it.
record_build() {
  local name="$1" commit="$2" digest="$3"
  echo "$name $commit $digest" >>"$work/builds"
  [ -f "$work/diff/$commit" ] || git diff --name-status --no-renames -z "$commit" HEAD >"$work/diff/$commit"
}

case "$mode" in
  reuse)
    [[ "$arg" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || usage
    commits="$(git rev-list --first-parent --max-count="$depth" HEAD)"
    while read -r name image; do
      for commit in $commits; do
        if digest="$(bash "$here/verified-digest.sh" "$image:sha-${commit:0:7}" "$arg")"; then
          record_build "$name" "$commit" "$digest"
          break
        fi
      done
    done < <(gojq -r '.[] | "\(.name) \(.image)"' "$work/matrix.json")
    ;;
  base)
    base="$(git rev-parse -q --verify "$arg^{commit}")" || {
      echo "ci-images-select: $arg is not a commit" >&2
      exit 1
    }
    while read -r name; do
      record_build "$name" "$base" -
    done < <(gojq -r '.[].name' "$work/matrix.json")
    ;;
  all) ;;
  *) usage ;;
esac

go run ./tools/affected images -in "$work" -modules "$modules" -mode "$mode"
