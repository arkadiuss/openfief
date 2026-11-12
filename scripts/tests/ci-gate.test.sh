#!/usr/bin/env bash
# The gate the pre-commit hook leaves out is CI's, and the downloads CI makes retry (docs/dev/pipeline.md#gate-tiers;
# kibi: TEST-tooling-ci-gate). Reads the real workflows, the setup action and the scripts: ci.yml runs on pushes to main and
# to the implementation branch, its check job runs the whole `task check`, only main publishes, no workflow installs
# cosign except through the retrying composite action, and every download of ours goes through scripts/retry.sh.
# Then a synthetic script with a bare download, which the download check must refuse. Needs gojq on PATH; offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
ci="$root/.github/workflows/ci.yml"

# check records one case: $2 is yes when it held.
check() {
  cases=$((cases + 1))
  if [ "$2" != yes ]; then
    echo "FAIL: $1" >&2
    failures=$((failures + 1))
  fi
}

# bare_downloads prints each line of the files that downloads or installs something without scripts/retry.sh: a curl
# with -f or -o (the flags of a download), a go install, an npm ci. A comment or a task's desc line doesn't count.
bare_downloads() {
  grep -HnE '(^|[^[:alnum:]_-])(curl .*-[a-zA-Z]*[fo]|go install|npm ci)([^[:alnum:]_-]|$)' "$@" | grep -v 'retry\.sh' | grep -vE '^[^:]*:[0-9]+:[[:space:]]*(#|desc:)' || true
}

# yes_if prints yes when the gojq filter $1 is true for the YAML file $2.
yes_if() {
  if [ "$(gojq -r --yaml-input "$1" "$2")" = true ]; then echo yes; else echo no; fi
}

check "ci.yml runs on pushes to main and to the implementation branch" \
  "$(yes_if '(.on // .true).push.branches | (index("main") != null) and (index("claude/openfief-implementation-*") != null)' "$ci")"
check "ci.yml still runs on pull requests to main" "$(yes_if '(.on // .true).pull_request.branches | index("main") != null' "$ci")"
check "the check job runs the whole task check, on every push" \
  "$(yes_if '.jobs.check | (has("if") | not) and any(.steps[]; .run == "task check")' "$ci")"
for job in docs-build docs deploy-demo; do
  check "the $job job runs only on main" "$(yes_if ".jobs[\"$job\"].if | contains(\"github.ref == 'refs/heads/main'\")" "$ci")"
done
check "images are pushed and signed only on main" \
  "$(yes_if '.jobs.images.with.push | contains("github.ref == '"'"'refs/heads/main'"'"'")' "$ci")"

# No workflow uses the cosign installer itself; the composite action does, three times, and only the first two may fail.
installer="$(grep -l 'sigstore/cosign-installer' "$root/.github/workflows"/*.yml || true)"
check "no workflow uses sigstore/cosign-installer directly" "$([ -z "$installer" ] && echo yes || echo no)"
composite="$root/.github/actions/install-cosign/action.yml"
check "the install-cosign action tries the installer three times, tolerating the first two failures" \
  "$(yes_if '[.runs.steps[] | select((.uses // "") | startswith("sigstore/cosign-installer@"))] | (length == 3) and (map(.["continue-on-error"] == true) == [true, true, false])' "$composite")"
check "the install-cosign action's three installers are the same pinned commit" \
  "$(yes_if '[.runs.steps[] | select((.uses // "") | startswith("sigstore/cosign-installer@")) | .uses] | (unique | length) == 1 and (.[0] | test("@[0-9a-f]{40}$"))' "$composite")"
users="$(grep -c 'uses: ./.github/actions/install-cosign' "$root/.github/workflows"/*.yml | awk -F: '{ n += $2 } END { print n }')"
check "eight workflow steps install cosign through the action (images 3, release 3, release-smoke, deploy-demo)" "$([ "$users" -eq 8 ] && echo yes || echo no)"

# Every download of ours goes through retry.sh: the Taskfile's tool installs, the scripts, the setup action.
check "the Taskfile downloads only through retry.sh" "$([ -z "$(bare_downloads "$root/Taskfile.yml")" ] && echo yes || echo no)"
check "the download scripts use retry.sh" \
  "$([ -z "$(bare_downloads "$root/scripts/envtest-assets.sh" "$root/scripts/e2e-gateway.sh")" ] && echo yes || echo no)"
check "the setup action installs Task through retry.sh" "$([ -z "$(bare_downloads "$root/.github/actions/setup/action.yml")" ] && echo yes || echo no)"
check "the setup action caches the pinned tools by their versions, never the e2e state" \
  "$(yes_if '.runs.steps[] | select((.uses // "") | startswith("actions/cache@")) | (.with.key | contains("tools-key")) and (.with.path | contains("!.bin/e2e"))' "$root/.github/actions/setup/action.yml")"

# The download check refuses what it exists to catch.
cat >"$work/bare.sh" <<'EOF'
# curl -fsSL in a comment is nothing
bash retry.sh curl -fsSL -o a https://example.com/a
curl -fsSL -o b https://example.com/b
GOBIN=x go install example.com/tool@v1.0.0
npm ci
EOF
check "a bare curl, go install and npm ci are each found, and a retried one and a comment aren't" \
  "$([ "$(bare_downloads "$work/bare.sh" | wc -l | tr -d ' ')" = 3 ] && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "ci-gate.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "ci-gate.test.sh: all $cases checks passed"
