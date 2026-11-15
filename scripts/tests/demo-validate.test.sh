#!/usr/bin/env bash
# Tests for scripts/demo/validate.sh on copies of deploy/ (kibi: TEST-tooling-demo-validate).
# Needs kustomize and gojq on PATH (task test adds .bin); fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0

# fresh copies deploy/ into a throwaway repository root, so Flux paths (./deploy/...) resolve the same way.
fresh() {
  rm -rf "$work/repo"
  mkdir -p "$work/repo"
  cp -r "$root/deploy" "$work/repo/deploy"
}

# validate runs validate.sh in the throwaway root; output lands in $work/out.
validate() {
  (cd "$work/repo" && bash "$root/scripts/demo/validate.sh") >"$work/out" 2>&1
}

# expect records whether validate passed or failed as wanted, and that the output names the problem.
expect() {
  local name="$1" want="$2" pattern="$3" got=pass
  cases=$((cases + 1))
  validate || got=fail
  if [ "$got" != "$want" ] || ! grep -q "$pattern" "$work/out"; then
    echo "FAIL: $name: want $want with '$pattern', got $got:" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

fresh
expect "the committed manifests" pass "every image is pinned by digest"

fresh
sed -i 's/    provider: cosign/    provider: notation/' "$work/repo/deploy/demo/stack.yaml"
expect "an OCI source without cosign verification" fail "OCI sources without cosign verification: openfief-demo-app"

fresh
sed -i '/^images:/,$d' "$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "an image pinned by tag only" fail "not pinned by digest: ghcr.io/arkadiuss/openfief/"

fresh
sed -i 's#path: ./deploy/demo/database#path: ./deploy/demo/missing#' "$work/repo/deploy/demo/stack.yaml"
expect "a Flux path that does not build" fail "kustomize build ./deploy/demo/missing failed"

fresh
sed -i 's/[$]{DEMO_HOST}/${DEMO_HOSTNAME}/' "$work/repo/deploy/demo/routes/openfief.yaml"
expect "a variable the demo settings don't define" fail "uses variables the demo settings don't define: DEMO_HOSTNAME"

# The model layer (deploy/demo/model) is ready and off: a Component the live demo does not list.
fresh
sed -i 's#"openrouter/google/gemma-4-26b-a4b-it"#"openrouter/openai/gpt-oss-120b"#' "$work/repo/deploy/demo/model/settings.json"
expect "a route to a model the provider instance does not list" fail "name openrouter/openai/gpt-oss-120b, which the instance openrouter does not list"

fresh
sed -i 's#"routes": {#"routes": {\n    "maintainer": ["openrouter/google/gemma-4-26b-a4b-it"],#' "$work/repo/deploy/demo/model/settings.json"
expect "a route for the maintainer" fail "route something other than the Ask box"

fresh
printf '\ncomponents:\n  - ../model\n' >>"$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "the model layer listed by the live demo" fail "the live demo (deploy/demo/openfief) includes the model layer"

fresh
rm "$work/repo/deploy/demo/model/openrouter.yaml"
expect "a model layer that does not build" fail "the model layer does not build with the demo"

# D-208's guards in the model layer: the core's flags, the UI's config and its grant.
fresh
sed -i 's/value: --ask-concurrent-calls=3/value: --ask-concurrent-calls=30/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "calls at once that are not D-208's" fail "the core's flags lack --ask-concurrent-calls=3"

fresh
sed -i 's/value: --ask-questions-per-hour=10/value: --ask-questions-per-hour=100/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "an hourly limit that is not D-208's" fail "the core's flags lack --ask-questions-per-hour=10"

fresh
sed -i 's/value: --ask-questions-per-day=30/value: --ask-questions-per-day=300/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a daily limit that is not D-208's" fail "the core's flags lack --ask-questions-per-day=30"

fresh
sed -i 's/value: --allow-turnstile/value: --allow-turnstile-not/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "no place for Turnstile in the core's policy" fail "the core's flags lack --allow-turnstile"

fresh
sed -i 's/value: --trusted-proxies=10.42.0.0\/16/value: --trusted-proxies-not/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "no trusted proxy for the address header" fail "the core's flags lack --trusted-proxies"

fresh
sed -i 's/value: --model-daily-budget-usd=1.00/value: --model-daily-budget-usd=lots/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a budget that is no amount" fail "the core's --model-daily-budget-usd is not a positive amount"

fresh
sed -i 's/value: --model-daily-budget-usd=1.00/value: --model-daily-budget-usd=0/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a budget of nothing" fail "the core's --model-daily-budget-usd is not a positive amount"

fresh
sed -i 's/value: \[challenges.cloudflare.com\]/value: [elsewhere.example.com]/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a grant that does not approve Turnstile's host" fail "the UI grant does not approve the host of turnstileVerifyUrl"

fresh
sed -i 's/value: \[ui--turnstile-secret\]/value: [ui--other]/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a grant that does not name the Turnstile Secret" fail "the UI grant does not name the Secret ui--turnstile-secret"

fresh
sed -i 's/askOutputTokens: 800/askOutputTokens: 1600/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "limits of one question that are not D-208's" fail "the UI's limits of one question are not D-208's"

fresh
printf '\ncomponents:\n  - ../model\n' >>"$work/repo/deploy/demo/openfief/kustomization.yaml"
expect "the placeholder site key in a live demo" fail "has the placeholder Turnstile site key"

fresh
sed -i 's/REPLACE-WITH-THE-TURNSTILE-SITE-KEY/0x4AAAAAAAxxxxxxxxxxxxxx/' "$work/repo/deploy/demo/model/kustomization.yaml"
expect "a model layer with a site key" pass "every image is pinned by digest"

if [ "$failures" -ne 0 ]; then
  echo "demo-validate: $failures of $cases cases failed" >&2
  exit 1
fi
echo "demo-validate: $cases cases passed"
