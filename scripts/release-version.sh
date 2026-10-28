#!/usr/bin/env bash
# Validates a release version for its channel (docs/dev/sdlc.md#releases).
# stable: YYYY.M.N; beta: YYYY.M.N-beta.K. No v prefix, no leading zeros, and the tag must not exist yet.
# Usage: release-version.sh <beta|stable> <version>
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: release-version.sh <beta|stable> <version>" >&2
  exit 2
fi

# valid_version succeeds when the version has the channel's calendar-version format.
valid_version() {
  local core='20[0-9]{2}\.(1[0-2]|[1-9])\.(0|[1-9][0-9]*)'
  case "$1" in
    stable) [[ "$2" =~ ^${core}$ ]] ;;
    beta) [[ "$2" =~ ^${core}-beta\.[1-9][0-9]*$ ]] ;;
    *) return 1 ;;
  esac
}

channel="$1"
version="$2"
case "$channel" in
  stable) format='YYYY.M.N' ;;
  beta) format='YYYY.M.N-beta.K' ;;
  *)
    echo "release: unknown channel '$channel' (beta or stable)" >&2
    exit 1
    ;;
esac
if ! valid_version "$channel" "$version"; then
  echo "release: '$version' is not a $channel version ($format, no v prefix, no leading zeros)" >&2
  exit 1
fi
if git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
  echo "release: tag '$version' already exists" >&2
  exit 1
fi
echo "release: $channel $version is valid"
