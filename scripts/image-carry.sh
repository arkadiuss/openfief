#!/usr/bin/env bash
# Carries an unchanged image forward to a new commit on main (docs/dev/pipeline.md#images). tools/affected found
# the digest signed for an earlier commit and saw nothing that goes into the image change since. This script
# checks that the digest's keyless signature comes from this repository's images workflow on main, tags the
# digest sha-<commit> and main without changing it (crane tag), and checks that sha-<commit> now verifies to the
# same digest, as scripts/demo/pin-images.sh will. So the demo only ever deploys a digest images.yml signed.
# Usage: image-carry.sh <image> <digest> <commit> <owner/repo>
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"

carry() {
  local image="${1:-}" digest="${2:-}" commit="${3:-}" repo="${4:-}" tag signed tagged
  if [[ ! "$image" =~ ^[a-z0-9][a-z0-9./_-]*$ ]] || [[ ! "$digest" =~ ^sha256:[0-9a-f]{64}$ ]] ||
    [[ ! "$commit" =~ ^[0-9a-f]{7,40}$ ]] || [[ ! "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
    echo "image-carry: usage: image-carry.sh <image> <sha256:digest> <commit> <owner/repo>" >&2
    return 1
  fi
  tag="sha-${commit:0:7}"
  if ! signed="$(bash "$here/verified-digest.sh" "$image@$digest" "$repo")" || [ "$signed" != "$digest" ]; then
    echo "image-carry: $image@$digest has no signature from $repo's images workflow on main; not carried" >&2
    return 1
  fi
  crane tag "$image@$digest" "$tag"
  crane tag "$image@$digest" main
  tagged="$(bash "$here/verified-digest.sh" "$image:$tag" "$repo")" || tagged="nothing verified"
  if [ "$tagged" != "$digest" ]; then
    echo "image-carry: $image:$tag verifies as $tagged, not the carried $digest" >&2
    return 1
  fi
  echo "image-carry: $image:$tag and :main -> $digest, signed by $repo's images workflow on main"
}

carry "$@"
