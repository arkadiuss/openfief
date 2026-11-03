#!/usr/bin/env bash
# Every reusable workflow that signs runs its signing jobs only in OpenFief's own repository (D-181;
# kibi: TEST-tooling-reusable-signing-guard). Any repository can call a workflow with a workflow_call trigger, and in
# a called workflow github.repository is the caller's, so a job that may request an OIDC token (id-token: write)
# must have an `if` that requires github.repository == 'arkadiuss/openfief' as a conjunct, not inside an `||`.
# Checks the real workflows, then synthetic ones the check must refuse. Needs gojq on PATH; fully offline.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
guard="github.repository == 'arkadiuss/openfief'"

# unguarded prints "<file>: <job>" for each job of the workflow file $1 that may request an OIDC token, in a workflow
# another repository can call, whose `if` doesn't require the repository guard. A job inherits the workflow's
# permissions when it sets none of its own. The guard counts only as a conjunct of the top-level `&&` chain.
unguarded() {
  gojq -r --yaml-input --arg guard "$guard" --arg file "$(basename "$1")" '
    def callable: (.on // .true) as $on
      | ($on | type) as $t
      | if $t == "string" then $on == "workflow_call"
        elif $t == "array" then any($on[]; . == "workflow_call")
        elif $t == "object" then $on | has("workflow_call")
        else false end;
    def oidc($p): $p == "write-all" or (($p | type) == "object" and $p["id-token"] == "write");
    def conjuncts: gsub("^\\s*\\$\\{\\{\\s*|\\s*\\}\\}\\s*$"; "") | [splits("\\s*&&\\s*")]
      | map(gsub("^\\s*\\(?\\s*|\\s*\\)?\\s*$"; ""));
    select(callable) | . as $wf
    | .jobs // {} | to_entries[]
    | select(oidc(.value.permissions // $wf.permissions))
    | select(((.value.if // "") | tostring | conjuncts | index($guard)) == null)
    | "\($file): \(.key)"' "$1"
}

# check records one case: the guard check of the workflow files $3... must find exactly $2 (empty for none).
check() {
  local name="$1" want="$2" got="" f
  shift 2
  cases=$((cases + 1))
  for f in "$@"; do
    got+="$(unguarded "$f")"$'\n'
  done
  got="$(sed '/^$/d' <<<"$got")"
  if [ "$got" != "$want" ]; then
    echo "FAIL: $name: want '$want', got '$got'" >&2
    failures=$((failures + 1))
  fi
}

# The real workflows: every signing job of a reusable workflow is guarded.
check "the repository's workflows" "" "$root/.github/workflows"/*.yml
# And the guard check sees them: images.yml and deploy-demo.yml are reusable and sign.
signing="$(gojq -r --yaml-input '
  select((.on // .true) | type == "object" and has("workflow_call")) | . as $wf
  | [.jobs[] | select((.permissions // $wf.permissions)["id-token"]? == "write")] | select(length > 0)
  | input_filename' "$root/.github/workflows"/*.yml | xargs -n1 basename | sort | tr '\n' ' ')"
cases=$((cases + 1))
if [ "$signing" != "deploy-demo.yml images.yml " ]; then
  echo "FAIL: the reusable signing workflows are $signing, want deploy-demo.yml and images.yml" >&2
  failures=$((failures + 1))
fi

write() { printf '%s\n' "$2" >"$work/$1"; }

write unguarded.yml 'on:
  workflow_call: {}
jobs:
  sign:
    runs-on: ubuntu-24.04
    permissions:
      id-token: write
    steps: [{run: "true"}]'
check "a reusable workflow's signing job without the guard" "unguarded.yml: sign" "$work/unguarded.yml"

write or.yml "on: [push, workflow_call]
jobs:
  sign:
    if: \${{ inputs.push || $guard }}
    runs-on: ubuntu-24.04
    permissions:
      id-token: write
    steps: [{run: 'true'}]"
check "the guard inside an or" "or.yml: sign" "$work/or.yml"

write other.yml "on: workflow_call
jobs:
  sign:
    if: github.repository == 'example/openfief'
    runs-on: ubuntu-24.04
    permissions: write-all
    steps: [{run: 'true'}]"
check "another repository, write-all permissions" "other.yml: sign" "$work/other.yml"

write inherited.yml "on:
  workflow_call: {}
permissions:
  id-token: write
jobs:
  sign:
    runs-on: ubuntu-24.04
    steps: [{run: 'true'}]
  quiet:
    runs-on: ubuntu-24.04
    permissions: {}
    steps: [{run: 'true'}]"
check "a job inheriting the workflow's id-token permission" "inherited.yml: sign" "$work/inherited.yml"

write guarded.yml "on:
  workflow_call: {}
jobs:
  sign:
    if: $guard && needs.discover.outputs.images != '[]'
    runs-on: ubuntu-24.04
    permissions:
      id-token: write
    steps: [{run: 'true'}]
  sign2:
    if: \${{ (needs.a.result == 'success') && $guard }}
    runs-on: ubuntu-24.04
    permissions:
      id-token: write
    steps: [{run: 'true'}]
  build:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
    steps: [{run: 'true'}]"
check "guarded signing jobs and a job that doesn't sign" "" "$work/guarded.yml"

write notreusable.yml 'on:
  push: {}
jobs:
  sign:
    runs-on: ubuntu-24.04
    permissions:
      id-token: write
    steps: [{run: "true"}]'
check "a workflow no other repository can call" "" "$work/notreusable.yml"

if [ "$failures" -gt 0 ]; then
  echo "reusable-signing-guard: $failures of $cases cases failed" >&2
  exit 1
fi
echo "reusable-signing-guard: $cases cases passed"
