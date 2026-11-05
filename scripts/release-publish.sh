#!/usr/bin/env bash
# Publishes a release from its release commit, HEAD (docs/dev/pipeline.md#releases; task release:publish, CI only):
# the tag at HEAD through the GitHub API, the GitHub release with its assets (a pre-release unless the channel is
# stable), and the pull request that brings the catalog commit on release/<version> back to main (D-88).
# Each step is skipped when an earlier attempt already did it, so re-running a publish that failed part-way finishes
# what is missing: a tag already at HEAD is kept (one anywhere else is refused), an existing release only gets the
# assets it lacks, by name, and an open or merged pull request from release/<version> is not opened again.
# Run from the repository root after task release:bundle, release:cli and release:assets, with the notes in
# dist/release-notes.md; gh reads GH_TOKEN and GH_REPO.
# Usage: release-publish.sh <alpha|beta|stable> <version> <cnpg-version>
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: release-publish.sh <alpha|beta|stable> <version> <cnpg-version>" >&2
  exit 2
fi
channel="$1"
version="$2"
cnpg_version="$3"

# release_assets prints the files every channel ships, one a line: the bundle, the docs and the signed CLI.
release_assets() {
  local f
  for f in "dist/bundle/openfief-crds-$version.yaml" "dist/bundle/openfief-install-$version.yaml" \
    "dist/bundle/openfief-db-$version.yaml" "dist/bundle/cnpg-$cnpg_version.yaml" \
    "dist/openfief-docs-$version.tar.gz" dist/release-notes.md; do
    [ -f "$f" ] || {
      echo "release:publish: $f is missing (task release:bundle, release:assets, release:notes)" >&2
      return 1
    }
  done
  [ -n "$(find dist/cli -type f 2>/dev/null)" ] || {
    echo "release:publish: dist/cli is empty (task release:cli)" >&2
    return 1
  }
  printf '%s\n' "dist/bundle/openfief-crds-$version.yaml" "dist/bundle/openfief-db-$version.yaml" \
    "dist/bundle/openfief-install-$version.yaml" "dist/bundle/cnpg-$cnpg_version.yaml" \
    "dist/openfief-docs-$version.tar.gz"
  find dist/cli -type f | LC_ALL=C sort
}

# ensure_tag creates the tag at HEAD unless it is already there; a tag at another commit is refused.
ensure_tag() {
  local head tagged
  head="$(git rev-parse HEAD)"
  if tagged="$(git rev-parse -q --verify "refs/tags/$version^{commit}")"; then
    if [ "$tagged" != "$head" ]; then
      echo "release:publish: tag $version already exists at $tagged, not at the release commit $head" >&2
      return 1
    fi
    echo "release:publish: tag $version already points at the release commit, kept"
    return 0
  fi
  gh api 'repos/{owner}/{repo}/git/refs' -f ref="refs/tags/$version" -f sha="$head" >/dev/null
  echo "release:publish: tagged $head as $version"
}

# ensure_release creates the GitHub release with every asset, or uploads to an existing one the assets it lacks.
# An asset whose upload never finished (state other than uploaded) counts as missing and is replaced.
ensure_release() {
  local assets=("$@") uploaded err missing=() f prerelease=()
  [ "$channel" = stable ] || prerelease=(--prerelease)
  err="$(mktemp)"
  if uploaded="$(gh release view "$version" --json assets --jq '.assets[] | select(.state == "uploaded") | .name' 2>"$err")"; then
    rm -f "$err"
    for f in "${assets[@]}"; do
      grep -qxF "$(basename "$f")" <<<"$uploaded" || missing+=("$f")
    done
    if [ "${#missing[@]}" -eq 0 ]; then
      echo "release:publish: release $version already has all ${#assets[@]} assets"
      return 0
    fi
    gh release upload "$version" --clobber "${missing[@]}"
    echo "release:publish: release $version existed; uploaded the ${#missing[@]} missing assets"
    return 0
  fi
  if ! grep -q 'release not found' "$err"; then
    cat "$err" >&2
    rm -f "$err"
    echo "release:publish: can't read release $version" >&2
    return 1
  fi
  rm -f "$err"
  gh release create "$version" --verify-tag "${prerelease[@]}" --title "$version" --notes-file dist/release-notes.md "${assets[@]}"
  echo "release:publish: created release $version with ${#assets[@]} assets"
}

# ensure_pull_request opens the pull request from release/<version> to main unless an open or merged one exists.
ensure_pull_request() {
  local existing
  existing="$(gh pr list --base main --head "release/$version" --state all --json number,state \
    --jq '.[] | select(.state == "OPEN" or .state == "MERGED") | "#\(.number) (\(.state))"')"
  if [ -n "$existing" ]; then
    echo "release:publish: pull request from release/$version already exists: $existing"
    return 0
  fi
  gh pr create --base main --head "release/$version" --title "chore(release): catalog the first-party plugins of $version" \
    --body "Adds the plugin digests release $version signed to the catalog (D-88), as tagged in $version."
}

list="$(release_assets)"
mapfile -t assets <<<"$list"
ensure_tag
ensure_release "${assets[@]}"
ensure_pull_request
