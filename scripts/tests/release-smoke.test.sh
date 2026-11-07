#!/usr/bin/env bash
# Tests for scripts/release-smoke.sh with stand-ins for gh, cosign, kubectl and node, and a fake release whose CLI
# is a script (kibi: TEST-tooling-release-smoke): a release that verifies installs in the guide's order and passes;
# a wrong identity accepted, a bad checksum, an unpinned image, a CloudNativePG manifest that isn't the pinned one, a
# CLI reporting another version, a missing enrolment link, a failed passkey and a failed egress audit each fail it. Also checks that the
# script runs the guide's commands (docs/guides/install.md) and that the workflow calls it. Offline: no registry,
# Sigstore, cluster or browser.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
repo=example/openfief
version=2026.11.0-alpha.2
commit=0123456789abcdef0123456789abcdef01234567
cli_id="https://github.com/$repo/.github/workflows/release.yml@refs/heads/main"
images_id="https://github.com/$repo/.github/workflows/images.yml@refs/heads/main"
op="ghcr.io/$repo/operator@sha256:$(printf '1%.0s' {1..64})"
eg="ghcr.io/$repo/egress-proxy@sha256:$(printf '2%.0s' {1..64})"
co="ghcr.io/$repo/core@sha256:$(printf '3%.0s' {1..64})"
mkdir -p "$work/bin"

# The fake CLI, shipped as the release's linux-amd64 binary: version prints FAKE_CLI_VERSION, plugin install logs,
# enroll prints FAKE_LINK (when set) and runs until SIGTERM, exiting FAKE_ENROLL_EXIT.
cat >"$work/openfief" <<'EOF'
#!/usr/bin/env bash
echo "openfief $*" >>"$FAKE_LOG"
case "$1" in
  version) echo "$FAKE_CLI_VERSION" ;;
  plugin) echo "Installed ui." ;;
  enroll)
    trap 'exit "${FAKE_ENROLL_EXIT:-0}"' TERM
    [ -z "${FAKE_LINK:-}" ] || echo "$FAKE_LINK"
    while :; do sleep 0.1; done
    ;;
  *) exit 2 ;;
esac
EOF

# release makes the fake release's assets in $work/release: the bundle, with $1 as the install's images, and the
# CLI, checksums and SBOM, each with a bundle the fake cosign accepts for release.yml on main.
release() {
  local r="$work/release" f
  rm -rf "$r"
  mkdir -p "$r"
  printf 'kind: Namespace\nmetadata:\n  name: openfief\n' >"$r/openfief-crds-$version.yaml"
  printf 'kind: Cluster\nmetadata:\n  name: openfief-db\n' >"$r/openfief-db-$version.yaml"
  printf 'kind: Namespace\nmetadata:\n  name: cnpg-system\n' >"$r/cnpg-1.30.1.yaml"
  : >"$r/openfief-install-$version.yaml"
  for f in $1; do
    printf 'kind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - name: c\n          image: %s\n' "$f" >>"$r/openfief-install-$version.yaml"
  done
  cp "$work/openfief" "$r/openfief-$version-linux-amd64"
  echo '{"bomFormat":"CycloneDX"}' >"$r/openfief-$version.cdx.json"
  (cd "$r" && sha256sum "openfief-$version-linux-amd64" "openfief-$version.cdx.json" >"openfief-$version-SHA256SUMS")
  for f in "openfief-$version-linux-amd64" "openfief-$version.cdx.json" "openfief-$version-SHA256SUMS"; do
    echo "$cli_id $repo refs/heads/main $(sha256sum <"$r/$f" | cut -d' ' -f1)" >"$r/$f.sigstore.json"
  done
}

# gh copies each --pattern asset from $work/release into the current directory and answers the tag's commit.
cat >"$work/bin/gh" <<'EOF'
#!/usr/bin/env bash
echo "gh $*" >>"$FAKE_LOG"
case "$1 $2" in
  "release download")
    shift 2
    while [ $# -gt 0 ]; do
      [ "$1" = --pattern ] && { [ ! -f "$FAKE_RELEASE/$2" ] || cp "$FAKE_RELEASE/$2" .; }
      shift
    done
    ;;
  "api repos/"*) echo "$FAKE_COMMIT" ;;
  *) exit 2 ;;
esac
EOF
# cosign verify-blob passes when the bundle holds "<identity> <repository> <ref> <sha256>" for what it was given;
# verify passes for the images in FAKE_SIGNED as images.yml on main. FAKE_LAX=1 accepts any identity.
cat >"$work/bin/cosign" <<'EOF'
#!/usr/bin/env bash
echo "cosign $*" >>"$FAKE_LOG"
cmd="$1"
shift
bundle="" id="" iss="" repo="" ref="" target=""
while [ $# -gt 0 ]; do
  case "$1" in
    --bundle) bundle="$2"; shift 2 ;;
    --certificate-identity) id="$2"; shift 2 ;;
    --certificate-oidc-issuer) iss="$2"; shift 2 ;;
    --certificate-github-workflow-repository) repo="$2"; shift 2 ;;
    --certificate-github-workflow-ref) ref="$2"; shift 2 ;;
    *) target="$1"; shift ;;
  esac
done
[ "$iss" = https://token.actions.githubusercontent.com ] && [ -n "$repo" ] && [ -n "$ref" ] || exit 1
case "$cmd" in
  verify-blob)
    want="$(cat "$bundle")"
    [ "${FAKE_LAX:-0}" = 1 ] && id="${want%% *}"
    [ "$want" = "$id $repo $ref $(sha256sum <"$target" | cut -d' ' -f1)" ]
    ;;
  verify)
    [ "${FAKE_LAX:-0}" = 1 ] || [ "$id" = "https://github.com/$repo/.github/workflows/images.yml@refs/heads/main" ] || exit 1
    case " $FAKE_SIGNED " in *" $target "*) echo '[{}]' ;; *) exit 1 ;; esac
    ;;
  *) exit 2 ;;
esac
EOF
# kubectl logs its arguments and fails when they match FAKE_KUBECTL_FAIL.
cat >"$work/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
echo "kubectl $*" >>"$FAKE_LOG"
[ -z "${FAKE_KUBECTL_FAIL:-}" ] || case "$*" in *$FAKE_KUBECTL_FAIL*) exit 1 ;; esac
EOF
# egress-proxy logs its arguments and exits FAKE_AUDIT_EXIT, as `egress-proxy audit` does on a finding.
cat >"$work/bin/egress-proxy" <<'EOF'
#!/usr/bin/env bash
echo "egress-proxy $*" >>"$FAKE_LOG"
exit "${FAKE_AUDIT_EXIT:-0}"
EOF
# node logs the link it was given and exits FAKE_NODE_EXIT.
cat >"$work/bin/node" <<'EOF'
#!/usr/bin/env bash
echo "node $(basename "$1") link=$OPENFIEF_E2E_ENROLL_LINK" >>"$FAKE_LOG"
exit "${FAKE_NODE_EXIT:-0}"
EOF
chmod +x "$work/bin/"* "$work/openfief"
cnpg_sum() { sha256sum <"$work/release/cnpg-1.30.1.yaml" | cut -d' ' -f1; }

# run runs release-smoke.sh in a new work directory with the fakes; its output lands in $work/out, the calls in
# $work/log. Extra environment (VAR=value) comes first.
run() {
  : >"$work/log"
  rm -rf "$work/w"
  env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_RELEASE="$work/release" FAKE_COMMIT="$commit" \
    FAKE_SIGNED="$op $eg $co" FAKE_CLI_VERSION="openfief $version commit $commit" \
    FAKE_LINK='http://localhost:8080/enroll#code=ofenroll_example' LINK_TIMEOUT=5 \
    CNPG_VERSION=1.30.1 CNPG_SHA256="$(cnpg_sum)" PASSKEY="$root/test/e2e/passkey.mjs" EGRESS_AUDIT="$work/bin/egress-proxy" \
    "$@" bash "$root/scripts/release-smoke.sh" "$version" "$repo" "$work/w" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the last run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}
yes_if() { if "$@"; then echo yes; else echo no; fi; }
passes() { if run "$@"; then echo yes; else echo no; fi; }
# fails_with runs with the environment after the pattern and wants a failure whose output matches the pattern.
fails_with() {
  local pattern="$1"
  shift
  if run "$@"; then echo no; elif grep -q -- "$pattern" "$work/out"; then echo yes; else echo no; fi
}

release "$op $eg $co"
check "a release that verifies installs" "$(passes)"
# The steps in the guide's order: download, CLI, images, CloudNativePG, CRDs, database, install, UI, enrol.
order="$(grep -oE '^(gh release download|cosign verify-blob|cosign verify |kubectl apply --server-side -f [^ ]+|kubectl -n openfief wait [^ ]+|kubectl -n openfief rollout status [^ ]+|openfief plugin install ui --yes|openfief enroll|node passkey.mjs|egress-proxy audit)' "$work/log" | uniq | tr '\n' '|')"
want="gh release download|cosign verify-blob|cosign verify |kubectl apply --server-side -f cnpg-1.30.1.yaml|kubectl apply --server-side -f openfief-crds-$version.yaml|kubectl apply --server-side -f openfief-db-$version.yaml|kubectl -n openfief wait cluster.postgresql.cnpg.io/openfief-db|kubectl apply --server-side -f openfief-install-$version.yaml|kubectl -n openfief rollout status deployment/openfief-operator|kubectl -n openfief rollout status deployment/egress-proxy|kubectl -n openfief rollout status deployment/openfief-core|openfief plugin install ui --yes|kubectl -n openfief wait plugin.openfief.io/ui|openfief enroll|node passkey.mjs|egress-proxy audit|"
check "the steps run in the guide's order, then the egress audit" "$(yes_if test "$order" = "$want")"
check "it verifies the CLI, the checksums and the SBOM" "$(yes_if test "$(grep -c "^cosign verify-blob .*--certificate-identity $cli_id " "$work/log")" -eq 3)"
check "it tries the CLI under a wrong identity" "$(yes_if grep -q "^cosign verify-blob .*--certificate-identity $images_id .*openfief-$version-linux-amd64\$" "$work/log")"
check "it verifies each image once as images.yml" "$(yes_if test "$(grep -c "^cosign verify ghcr.io/.* --certificate-identity $images_id " "$work/log")" -eq 3)"
check "it tries an image under a wrong identity" "$(yes_if grep -q "^cosign verify ghcr.io/.* --certificate-identity $cli_id " "$work/log")"
check "the passkey script gets the enrolment link" "$(yes_if grep -qxF 'node passkey.mjs link=http://localhost:8080/enroll#code=ofenroll_example' "$work/log")"
check "the link stays out of the log" "$(yes_if test "$(grep -c ofenroll_example "$work/out")" -eq 0)"
mkdir -p "$work/full"
touch "$work/full/x"
if PATH="$work/bin:$PATH" CNPG_VERSION=1.30.1 CNPG_SHA256="$(cnpg_sum)" PASSKEY="$root/test/e2e/passkey.mjs" \
  EGRESS_AUDIT="$work/bin/egress-proxy" bash "$root/scripts/release-smoke.sh" "$version" "$repo" "$work/full" >"$work/out" 2>&1; then
  check "a work directory that isn't empty is refused" no
else
  check "a work directory that isn't empty is refused" "$(yes_if grep -q 'is not empty' "$work/out")"
fi

check "a wrong identity accepted fails" "$(fails_with 'a wrong identity' FAKE_LAX=1)"
check "a CLI reporting another version fails" "$(fails_with 'openfief version printed' FAKE_CLI_VERSION='openfief dev commit unknown')"
check "CLI_VERSION=0 skips the version check" "$(passes FAKE_CLI_VERSION='usage' CLI_VERSION=0)"
check "a missing link fails" "$(fails_with 'printed no link' FAKE_LINK=)"
check "a failed passkey fails" "$(fails_with 'first passkey' FAKE_NODE_EXIT=1)"
check "a failed egress audit fails" "$(fails_with "the egress proxy's log breaks" FAKE_AUDIT_EXIT=1)"
check "an enrol that doesn't stop cleanly fails" "$(fails_with 'openfief enroll exited with' FAKE_ENROLL_EXIT=3)"
check "a rollout that fails prints diagnostics" "$(fails_with 'release-smoke: diagnostics' FAKE_KUBECTL_FAIL='rollout status deployment/openfief-core')"
check "diagnostics never read a Secret" "$(yes_if test "$(grep -ci secret "$work/log")" -eq 0)"
check "a CloudNativePG manifest that isn't pinned fails before the cluster" "$(fails_with "doesn't match the pinned sha256" CNPG_SHA256="$(printf 'a%.0s' {1..64})")"
check "nothing touched the cluster then" "$(yes_if test "$(grep -c '^kubectl' "$work/log")" -eq 0)"

release "$op ghcr.io/$repo/core:main"
check "an image by tag fails" "$(fails_with 'not an image by digest')"
release "$op $eg $co"
echo tampered >>"$work/release/openfief-$version.cdx.json"
check "a tampered SBOM fails" "$(fails_with "signature doesn't verify")"
release "$op $eg"
check "an unsigned image fails" "$(fails_with "verify as $images_id" FAKE_SIGNED="$op")"

# The script runs the guide's commands, so the guide stays the thing it tests.
guide="$root/docs/guides/install.md"
smoke="$root/scripts/release-smoke.sh"
for line in \
  'sha256sum --check --ignore-missing' \
  "sed -n 's/^ *image: *//p'" \
  'kubectl apply --server-side -f' \
  'kubectl -n cnpg-system rollout status deployment/cnpg-controller-manager --timeout=5m' \
  'kubectl -n openfief wait cluster.postgresql.cnpg.io/openfief-db --for=condition=Ready --timeout=10m' \
  'for d in openfief-operator egress-proxy openfief-core; do' \
  'kubectl -n openfief rollout status "deployment/$d" --timeout=5m' \
  'openfief plugin install ui' \
  'kubectl -n openfief wait plugin.openfief.io/ui --for=condition=Ready --timeout=5m' \
  'openfief enroll' \
  'openfief version'; do
  check "the guide and the script both run: $line" "$(yes_if bash -c 'grep -qF -- "$1" "$2" && grep -qF -- "$1" "$3"' _ "$line" "$guide" "$smoke")"
done
workflow="$root/.github/workflows/release-smoke.yml"
check "the workflow runs task release:smoke" "$(yes_if grep -q 'run: task release:smoke' "$workflow")"
check "release.yml calls the smoke test after publishing" "$(yes_if grep -q 'uses: ./.github/workflows/release-smoke.yml' "$root/.github/workflows/release.yml")"

if [ "$failures" -ne 0 ]; then
  echo "release-smoke: $failures of $cases checks failed" >&2
  exit 1
fi
echo "release-smoke: $cases checks passed"
