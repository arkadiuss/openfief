#!/usr/bin/env bash
# Tests for scripts/demo/render-audit.sh on copies of deploy/demo (kibi: TEST-tooling-demo-render-audit).
# Needs kustomize and gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
digest='sha256:1111111111111111111111111111111111111111111111111111111111111111'

# fresh copies deploy/demo into the work directory, the egress proxy's digest pinned as pin-images.sh leaves it.
fresh() {
  rm -rf "$work/demo" "$work/out.yaml" "$work/log"
  cp -r "$root/deploy/demo" "$work/demo"
  awk -v digest="$digest" '
    $0 ~ "^  - name: ghcr.io/arkadiuss/openfief/egress-proxy$" { found = 1; print; next }
    found && $1 == "digest:" { sub(/digest: .*/, "digest: " digest); found = 0 }
    { print }' "$root/deploy/demo/openfief/kustomization.yaml" >"$work/demo/openfief/kustomization.yaml"
}

# render runs render-audit.sh with tag $1 on the copy; its output lands in $work/log.
render() {
  bash "$root/scripts/demo/render-audit.sh" "$1" "$work/out.yaml" "$work/demo" >"$work/log" 2>&1
}

# check records one case: the gojq filter $2 over the rendered manifests must print true.
check() {
  local name="$1" filter="$2" got
  cases=$((cases + 1))
  got="$(gojq --yaml-input -s "$filter" "$work/out.yaml" 2>&1)" || true
  if [ "$got" != "true" ]; then
    echo "FAIL: $name: $filter printed $got" >&2
    failures=$((failures + 1))
  fi
}

# refuses records a case where render-audit.sh must fail, naming the problem.
refuses() {
  local name="$1" tag="$2" pattern="$3"
  cases=$((cases + 1))
  if render "$tag" || ! grep -q "$pattern" "$work/log"; then
    echo "FAIL: $name: want a failure with '$pattern', got:" >&2
    sed 's/^/  | /' "$work/log" >&2
    failures=$((failures + 1))
  fi
}

job='[.[] | select(.kind == "Job" and .metadata.name == "egress-audit")] | .[0]'

fresh
render sha-1a2b3c4 || {
  echo "FAIL: render-audit.sh failed on a pinned copy:" >&2
  sed 's/^/  | /' "$work/log" >&2
  exit 1
}
check "the Job runs egress-proxy audit at the pinned digest" \
  "$job | .spec.template.spec.containers == [.spec.template.spec.containers[0]] and .spec.template.spec.containers[0].image == \"ghcr.io/arkadiuss/openfief/egress-proxy@$digest\" and .spec.template.spec.containers[0].args == [\"audit\"]"
check "the Job's template carries the deploy's tag, so each deploy re-creates it" \
  "$job | .spec.template.metadata.annotations[\"openfief.io/deploy\"] == \"sha-1a2b3c4\" and .metadata.annotations[\"kustomize.toolkit.fluxcd.io/force\"] == \"enabled\""
check "the Job runs once, as the audit's ServiceAccount, in openfief" \
  "$job | .metadata.namespace == \"openfief\" and .spec.backoffLimit == 0 and .spec.template.spec.serviceAccountName == \"egress-audit\""
check "the audit's Role only lists pods and grants and reads pod logs" \
  "[.[] | select(.kind == \"Role\" and .metadata.name == \"egress-audit\")] | .[0].rules == [{apiGroups: [\"\"], resources: [\"pods\"], verbs: [\"list\"]}, {apiGroups: [\"\"], resources: [\"pods/log\"], verbs: [\"get\"]}, {apiGroups: [\"openfief.io\"], resources: [\"plugingrants\"], verbs: [\"list\"]}]"
check "the rendered audit holds only the Job and its RBAC" \
  "[.[] | .kind] | sort == [\"Job\", \"Role\", \"RoleBinding\", \"ServiceAccount\"]"
cases=$((cases + 1))
if ! grep -q 'sha256:0000000000000000000000000000000000000000000000000000000000000000' "$work/demo/audit/kustomization.yaml"; then
  echo "FAIL: render-audit.sh changed the committed audit directory" >&2
  failures=$((failures + 1))
fi

fresh
cp "$root/deploy/demo/openfief/kustomization.yaml" "$work/demo/openfief/kustomization.yaml"
refuses "the egress proxy's placeholder digest" sha-1a2b3c4 "still the placeholder"

fresh
sed -i '/egress-proxy$/,+1d' "$work/demo/openfief/kustomization.yaml"
refuses "no digest for the egress proxy" sha-1a2b3c4 "pins no digest for ghcr.io/arkadiuss/openfief/egress-proxy"

fresh
refuses "a tag that isn't one" 'bad tag;' "is not an image tag"

if [ "$failures" -ne 0 ]; then
  echo "demo-render-audit: $failures of $cases cases failed" >&2
  exit 1
fi
echo "demo-render-audit: $cases cases passed"
