#!/usr/bin/env bash
# Tests for scripts/demo/pin-images.sh with a fake cosign, on a copy of deploy/demo; no registry or Sigstore call
# (kibi: TEST-tooling-demo-pin-images). Needs gojq and kustomize on PATH (task test adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
op='sha256:1111111111111111111111111111111111111111111111111111111111111111'
eg='sha256:2222222222222222222222222222222222222222222222222222222222222222'
co='sha256:3333333333333333333333333333333333333333333333333333333333333333'
ds='sha256:4444444444444444444444444444444444444444444444444444444444444444'
nf='sha256:5555555555555555555555555555555555555555555555555555555555555555'
ui='sha256:6666666666666666666666666666666666666666666666666666666666666666'

# The fake cosign logs its arguments and "verifies" only the images listed in FAKE_SIGNED (image:tag=digest).
mkdir -p "$work/bin"
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$FAKE_LOG"
[ "$1" = verify ] || exit 2
for entry in $FAKE_SIGNED; do
  if [ "${entry%%=*}" = "$2" ]; then
    printf '[{"critical":{"image":{"docker-manifest-digest":"%s"}}}]\n' "${entry#*=}"
    exit 0
  fi
done
echo "no matching signatures" >&2
exit 1
EOF
chmod +x "$work/bin/cosign"

# pin copies deploy/demo, runs pin-images.sh on it with FAKE_SIGNED, and leaves the copy in $work/demo.
pin() {
  rm -rf "$work/demo" "$work/log"
  cp -r "$root/deploy/demo" "$work/demo"
  : >"$work/log"
  PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_SIGNED="$1" \
    bash "$root/scripts/demo/pin-images.sh" "$2" example/openfief "$work/demo" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

# yes_if prints yes when the command succeeds.
yes_if() {
  if "$@" >/dev/null 2>&1; then echo yes; else echo no; fi
}

images="ghcr.io/arkadiuss/openfief/operator:sha-1a2b3c4=$op ghcr.io/arkadiuss/openfief/egress-proxy:sha-1a2b3c4=$eg"
images="$images ghcr.io/arkadiuss/openfief/core:sha-1a2b3c4=$co ghcr.io/arkadiuss/openfief/demo-seed:sha-1a2b3c4=$ds"
images="$images ghcr.io/arkadiuss/openfief/nbp-fx:sha-1a2b3c4=$nf ghcr.io/arkadiuss/openfief/ui:sha-1a2b3c4=$ui"
if pin "$images" sha-1a2b3c4; then r=yes; else r=no; fi
check "every image signed: pinning succeeds" "$r"
check "operator digest is pinned" "$(yes_if gojq --yaml-input -e --arg d "$op" '.images[] | select(.name | endswith("/operator")) | .digest == $d' "$work/demo/openfief/kustomization.yaml")"
check "egress-proxy digest is pinned" "$(yes_if gojq --yaml-input -e --arg d "$eg" '.images[] | select(.name | endswith("/egress-proxy")) | .digest == $d' "$work/demo/openfief/kustomization.yaml")"
check "core digest is pinned" "$(yes_if gojq --yaml-input -e --arg d "$co" '.images[] | select(.name | endswith("/core")) | .digest == $d' "$work/demo/openfief/kustomization.yaml")"
check "demo-seed digest is pinned" "$(yes_if gojq --yaml-input -e --arg d "$ds" '.images[] | select(.name | endswith("/demo-seed")) | .digest == $d' "$work/demo/openfief/kustomization.yaml")"
check "nbp-fx digest is pinned" "$(yes_if gojq --yaml-input -e --arg d "$nf" '.images[] | select(.name | endswith("/nbp-fx")) | .digest == $d' "$work/demo/openfief/kustomization.yaml")"
# Render the pinned kustomization with the rest of deploy/, which it builds on.
rm -rf "$work/deploy"
cp -r "$root/deploy" "$work/deploy"
cp "$work/demo/openfief/kustomization.yaml" "$work/deploy/demo/openfief/kustomization.yaml"
kustomize build "$work/deploy/demo/openfief" >"$work/rendered.yaml" 2>"$work/out" || true
check "the nbp Plugin runs the pinned plugin image" "$(yes_if gojq --yaml-input -e --arg i "ghcr.io/arkadiuss/openfief/nbp-fx@$nf" 'select(.kind == "Plugin" and .metadata.name == "nbp") | .spec.image == $i' "$work/rendered.yaml")"
check "the ui Plugin runs the pinned plugin image" "$(yes_if gojq --yaml-input -e --arg i "ghcr.io/arkadiuss/openfief/ui@$ui" 'select(.kind == "Plugin" and .metadata.name == "ui") | .spec.image == $i' "$work/rendered.yaml")"
check "comments survive" "$(yes_if grep -q '^# OpenFief on the demo machine' "$work/demo/openfief/kustomization.yaml")"
check "no placeholder is left" "$(yes_if bash -c "! grep -q sha256:0000 '$work/demo/openfief/kustomization.yaml'")"
check "only the images workflow on main is trusted" "$(yes_if grep -q -- '--certificate-identity https://github.com/example/openfief/.github/workflows/images.yml@refs/heads/main --certificate-oidc-issuer https://token.actions.githubusercontent.com' "$work/log")"
check "the tag is what gets verified" "$(yes_if grep -q '^verify ghcr.io/arkadiuss/openfief/operator:sha-1a2b3c4 ' "$work/log")"

if pin "ghcr.io/arkadiuss/openfief/operator:2026.11.0-beta.1=$op" 2026.11.0-beta.1; then r=no; else r=yes; fi
check "an unsigned image fails the pin" "$r"
check "the failure names the image" "$(yes_if grep -q 'no verified signature for ghcr.io/arkadiuss/openfief/egress-proxy:2026.11.0-beta.1' "$work/out")"
check "the unsigned image keeps its placeholder" "$(yes_if grep -q "digest: sha256:0000" "$work/demo/openfief/kustomization.yaml")"

if pin "$images" 'bad tag'; then r=no; else r=yes; fi
check "a malformed tag is refused" "$r"
check "a malformed tag calls nothing" "$(yes_if test ! -s "$work/log")"

if pin "ghcr.io/arkadiuss/openfief/operator:sha-1=sha256:short ghcr.io/arkadiuss/openfief/egress-proxy:sha-1=$eg" sha-1; then r=no; else r=yes; fi
check "a malformed digest is refused" "$r"

if [ "$failures" -ne 0 ]; then
  echo "demo-pin-images: $failures of $cases checks failed" >&2
  exit 1
fi
echo "demo-pin-images: $cases checks passed"
