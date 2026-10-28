#!/usr/bin/env bash
# Release notes in Markdown from the Conventional Commits since the previous release (docs/dev/sdlc.md).
# A stable release starts after the previous stable one; a beta after the previous release of either channel.
# Usage: release-notes.sh <version> [ref, default HEAD]
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "usage: release-notes.sh <version> [ref]" >&2
  exit 2
fi
version="$1"
ref="${2:-HEAD}"

# previous_release prints the newest release tag reachable from the ref that the notes start after.
previous_release() {
  local exclude=(--exclude "$version")
  [[ "$version" == *-beta.* ]] || exclude+=(--exclude '*-beta.*')
  git describe --tags --abbrev=0 --match '[0-9][0-9][0-9][0-9].*' "${exclude[@]}" "$ref" 2>/dev/null || true
}

# section prints the notes section a commit belongs in, from its Conventional Commit subject and body.
section() {
  local re='^([a-z]+)(\([^)]*\))?(!)?: '
  if [[ ! "$1" =~ $re ]]; then
    echo other
  elif [ -n "${BASH_REMATCH[3]}" ] || grep -qE '^BREAKING[ -]CHANGE: ' <<<"$2"; then
    echo breaking
  else
    case "${BASH_REMATCH[1]}" in
      feat) echo features ;;
      fix) echo fixes ;;
      perf) echo performance ;;
      *) echo other ;;
    esac
  fi
}

previous="$(previous_release)"
range="$ref"
[ -z "$previous" ] || range="$previous..$ref"

declare -A lines=()
for commit in $(git rev-list --no-merges "$range"); do
  subject="$(git log -1 --format=%s "$commit")"
  body="$(git log -1 --format=%b "$commit")"
  key="$(section "$subject" "$body")"
  lines[$key]+="- $subject ($(git rev-parse --short "$commit"))"$'\n'
done

echo "## $version"
echo
if [ -n "$previous" ]; then
  echo "Changes since $previous."
else
  echo "First release."
fi
for key in breaking features fixes performance other; do
  [ -n "${lines[$key]:-}" ] || continue
  case "$key" in
    breaking) title='Breaking changes' ;;
    features) title='Features' ;;
    fixes) title='Fixes' ;;
    performance) title='Performance' ;;
    *) title='Other changes' ;;
  esac
  printf '\n### %s\n\n%s' "$title" "${lines[$key]}"
done
