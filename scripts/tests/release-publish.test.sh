#!/usr/bin/env bash
# Tests for scripts/release-publish.sh with a stand-in gh in a throwaway repository (kibi: TEST-tooling-release-publish):
# a first publish creates the tag, the release with every asset and the pull request; a re-run after a failure keeps
# what an earlier attempt did and finishes only what is missing.
set -euo pipefail
# A git hook exports these for the outer repository; the throwaway repository below must not inherit them.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
command -v gojq >/dev/null || {
  echo "release-publish.test: gojq is missing (task test:script puts .bin first on PATH)" >&2
  exit 1
}

# The stand-in gh logs each call that writes, one line each, and answers reads from the state directory: release.json
# (absent: no release), prs.json (the pull requests of release/<version>) and view-error (any other failure). --jq
# runs through gojq, as gh runs its own jq, so the script's filters are tested too.
mkdir -p "$work/bin"
cat >"$work/bin/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
jq_filter=
args=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --jq) jq_filter="$2"; shift 2 ;;
    *) args+=("$1"); shift ;;
  esac
done
answer() { if [ -n "$jq_filter" ]; then gojq -r "$jq_filter" "$1"; else cat "$1"; fi; }
case "${args[0]} ${args[1]}" in
  "api repos/{owner}/{repo}/git/refs") echo "tag ${args[*]:2}" >>"$STATE/calls" ;;
  "release view")
    if [ -f "$STATE/view-error" ]; then echo "HTTP 502: Bad Gateway" >&2; exit 1; fi
    [ -f "$STATE/release.json" ] || { echo "release not found" >&2; exit 1; }
    answer "$STATE/release.json"
    ;;
  "release create" | "release upload" | "pr create") echo "${args[*]}" >>"$STATE/calls" ;;
  "pr list") answer "$STATE/prs.json" ;;
  *) echo "gh stand-in: unexpected call: ${args[*]}" >&2; exit 1 ;;
esac
EOF
chmod +x "$work/bin/gh"
export PATH="$work/bin:$PATH"

export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.invalid
export GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.invalid
version=2026.11.0-alpha.2
cnpg=1.30.1
failures=0

# setup makes a fresh repository with a release commit on top of main and every file the publish ships.
setup() {
  rm -rf "$work/repo" "$work/state"
  mkdir -p "$work/repo" "$work/state"
  export STATE="$work/state"
  echo '[]' >"$STATE/prs.json"
  cd "$work/repo"
  git init -q
  git commit -q --allow-empty -m "chore: base"
  git commit -q --allow-empty -m "chore(release): catalog the first-party plugins of $version (D-88)"
  mkdir -p dist/bundle dist/cli
  for f in "openfief-crds-$version.yaml" "openfief-install-$version.yaml" "openfief-db-$version.yaml" "cnpg-$cnpg.yaml"; do
    echo "# $f" >"dist/bundle/$f"
  done
  for f in "openfief-$version-linux-amd64" "openfief-$version-linux-amd64.sigstore.json" "openfief-$version-SHA256SUMS"; do
    echo "$f" >"dist/cli/$f"
  done
  echo docs >"dist/openfief-docs-$version.tar.gz"
  echo "## $version" >dist/release-notes.md
}

# release_with writes a release whose assets have the given names; a name ending in :starter never finished uploading.
release_with() {
  local name state items=()
  for name in "$@"; do
    state=uploaded
    case "$name" in *:starter) state=starter name="${name%:starter}" ;; esac
    items+=("{\"name\":\"$name\",\"state\":\"$state\"}")
  done
  local IFS=,
  echo "{\"assets\":[${items[*]}]}" >"$STATE/release.json"
}

# check compares the writes the publish made with the expected ones and its exit status with the expected one.
check() {
  local name="$1" want_status="$2" want_calls="$3" status=0 got_calls out
  out="$(bash "$root/scripts/release-publish.sh" "${channel:-alpha}" "$version" "$cnpg" 2>&1)" || status=$?
  got_calls="$(cat "$STATE/calls" 2>/dev/null || true)"
  local got_status=ok
  [ "$status" -eq 0 ] || got_status=fail
  if [ "$got_status" != "$want_status" ]; then
    printf 'FAIL: %s: want %s, exit %s:\n%s\n' "$name" "$want_status" "$status" "$out" >&2
    failures=$((failures + 1))
  elif [ "$got_calls" != "$want_calls" ]; then
    printf 'FAIL: %s\nwant calls:\n%s\ngot calls:\n%s\noutput:\n%s\n' "$name" "$want_calls" "$got_calls" "$out" >&2
    failures=$((failures + 1))
  fi
}

all_assets="dist/bundle/openfief-crds-$version.yaml dist/bundle/openfief-db-$version.yaml dist/bundle/openfief-install-$version.yaml dist/bundle/cnpg-$cnpg.yaml dist/openfief-docs-$version.tar.gz dist/cli/openfief-$version-SHA256SUMS dist/cli/openfief-$version-linux-amd64 dist/cli/openfief-$version-linux-amd64.sigstore.json"
pr="pr create --base main --head release/$version --title chore(release): catalog the first-party plugins of $version --body Adds the plugin digests release $version signed to the catalog (D-88), as tagged in $version."

setup
check "a first publish tags HEAD, creates the pre-release with every asset and opens the pull request" ok \
  "tag -f ref=refs/tags/$version -f sha=$(git rev-parse HEAD)
release create $version --verify-tag --prerelease --title $version --notes-file dist/release-notes.md $all_assets
$pr"

setup
channel=stable check "a stable release is no pre-release" ok \
  "tag -f ref=refs/tags/$version -f sha=$(git rev-parse HEAD)
release create $version --verify-tag --title $version --notes-file dist/release-notes.md $all_assets
$pr"

# The alpha.1 case: tag and release done, the pull request failed.
setup
git tag "$version"
release_with "openfief-crds-$version.yaml" "openfief-db-$version.yaml" "openfief-install-$version.yaml" "cnpg-$cnpg.yaml" \
  "openfief-docs-$version.tar.gz" "openfief-$version-SHA256SUMS" "openfief-$version-linux-amd64" \
  "openfief-$version-linux-amd64.sigstore.json"
check "a re-run keeps the tag at HEAD and a complete release, and opens only the pull request" ok "$pr"

setup
git tag -a -m "$version" "$version"
release_with "openfief-crds-$version.yaml" "openfief-install-$version.yaml" "cnpg-$cnpg.yaml" \
  "openfief-$version-linux-amd64:starter" "openfief-$version-linux-amd64.sigstore.json"
echo '[{"number":7,"state":"OPEN"}]' >"$STATE/prs.json"
check "a re-run uploads only the missing assets, replaces an unfinished one, and opens no second pull request" ok \
  "release upload $version --clobber dist/bundle/openfief-db-$version.yaml dist/openfief-docs-$version.tar.gz dist/cli/openfief-$version-SHA256SUMS dist/cli/openfief-$version-linux-amd64"

setup
git tag "$version"
release_with "openfief-crds-$version.yaml" "openfief-db-$version.yaml" "openfief-install-$version.yaml" "cnpg-$cnpg.yaml" \
  "openfief-docs-$version.tar.gz" "openfief-$version-SHA256SUMS" "openfief-$version-linux-amd64" \
  "openfief-$version-linux-amd64.sigstore.json"
echo '[{"number":5,"state":"CLOSED"},{"number":7,"state":"MERGED"}]' >"$STATE/prs.json"
check "a re-run of a finished publish writes nothing" ok ""

setup
echo '[{"number":5,"state":"CLOSED"}]' >"$STATE/prs.json"
git tag "$version"
release_with "openfief-crds-$version.yaml" "openfief-db-$version.yaml" "openfief-install-$version.yaml" "cnpg-$cnpg.yaml" \
  "openfief-docs-$version.tar.gz" "openfief-$version-SHA256SUMS" "openfief-$version-linux-amd64" \
  "openfief-$version-linux-amd64.sigstore.json"
check "a closed, unmerged pull request doesn't count" ok "$pr"

setup
git tag "$version" HEAD~1
check "a tag at another commit is refused before anything is written" fail ""

setup
touch "$STATE/view-error"
check "a release that can't be read is not created again" fail "tag -f ref=refs/tags/$version -f sha=$(git rev-parse HEAD)"

setup
rm "dist/bundle/openfief-db-$version.yaml"
check "a missing bundle file is refused before anything is written" fail ""

setup
rm -r dist/cli
check "a missing CLI is refused before anything is written" fail ""

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-publish: 10 cases passed"
