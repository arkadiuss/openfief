#!/usr/bin/env bash
# Carries an unchanged image forward to a new commit on main (docs/dev/pipeline.md#images). tools/affected found
# the digest signed for an earlier commit and saw nothing that goes into the image change since. This script
# checks that the digest's keyless signature comes from this repository's images workflow on main, tags the
# digest sha-<commit> and main without changing it (crane tag), and checks that sha-<commit> now verifies to the
# same digest, as scripts/demo/pin-images.sh will, asking again a few times while that tag verifies as nothing
# (IMAGE_CARRY_ATTEMPTS, default 5; IMAGE_CARRY_WAIT seconds, default 5, doubling). So the demo only ever deploys
# a digest images.yml signed.
# Usage: image-carry.sh <image> <digest> <commit> <owner/repo>
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"

carry() {
  local image="${1:-}" digest="${2:-}" commit="${3:-}" repo="${4:-}" tag signed tagged attempts wait try
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
  # The registry and the transparency log may answer for a tag written a moment ago before they agree with it
  # (CI run 43504031407: the digest verified, then its new tag verified as nothing a second later). So a tag that
  # verifies as nothing is asked again, waiting longer each time; a tag that verifies as another digest fails at
  # once, since waiting won't change what it points at.
  attempts="${IMAGE_CARRY_ATTEMPTS:-5}" wait="${IMAGE_CARRY_WAIT:-5}"
  for ((try = 1; ; try++)); do
    tagged="$(bash "$here/verified-digest.sh" "$image:$tag" "$repo")" || tagged="nothing verified"
    if [ "$tagged" = "$digest" ] || [ "$tagged" != "nothing verified" ] || [ "$try" -ge "$attempts" ]; then
      break
    fi
    echo "image-carry: $image:$tag verifies as nothing yet (try $try of $attempts); asking again in ${wait}s" >&2
    sleep "$wait"
    wait=$((wait * 2))
  done
  if [ "$tagged" != "$digest" ]; then
    echo "image-carry: $image:$tag verifies as $tagged, not the carried $digest" >&2
    return 1
  fi
  echo "image-carry: $image:$tag and :main -> $digest, signed by $repo's images workflow on main"
}

carry "$@"
