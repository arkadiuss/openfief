#!/usr/bin/env bash
# Prints the digest of <ref> (image:tag or image@digest) whose keyless cosign signature comes from <owner/repo>'s
# images workflow on main, the identity scripts/demo/pin-images.sh trusts. Fails, printing nothing, when nothing
# verifies or the signatures name several digests. Used by image-carry.sh and ci-images-select.sh; needs cosign
# and gojq on PATH.
# Usage: verified-digest.sh <ref> <owner/repo>
set -euo pipefail

verified_digest() {
  local ref="$1" repo="$2" digest
  digest="$(cosign verify "$ref" \
    --certificate-identity "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main" \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    --output json 2>/dev/null | gojq -r '[.[].critical.image["docker-manifest-digest"]] | unique | .[]')" || true
  [[ "$digest" =~ ^sha256:[0-9a-f]{64}$ ]] || return 1
  echo "$digest"
}

verified_digest "$@"
