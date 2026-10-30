#!/usr/bin/env bash
# Tests for scripts/check-commit.sh, the pre-commit hook's runner, in a throwaway git repository with a fake go
# (go list and go run ./tools/affected) and a fake task; tools/affected's own decisions are its Go tests
# (kibi: TEST-tooling-check-commit).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# A hook exports GIT_DIR and friends; the throwaway repository must not inherit them.
unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE GIT_PREFIX GIT_OBJECT_DIRECTORY
failures=0
cases=0

mkdir -p "$work/bin"
# The fake go answers go list with an empty listing, and go run with the plan in FAKE_PLAN (or fails with
# FAKE_TOOL_FAILS), copying the input directory it was given to $work/seen.
cat >"$work/bin/go" <<'EOF'
#!/usr/bin/env bash
echo "go $*" >>"$FAKE_LOG"
case "$1" in
  list) echo '{}' ;;
  run)
    [ -z "${FAKE_TOOL_FAILS:-}" ] || exit 1
    shift 2
    while [ "$#" -gt 0 ]; do
      if [ "$1" = -in ]; then rm -rf "$FAKE_SEEN" && cp -r "$2" "$FAKE_SEEN"; fi
      shift
    done
    printf '%b' "$FAKE_PLAN"
    ;;
  *) exit 2 ;;
esac
EOF
# The fake task logs each call as one line of arguments joined by |, reads its standard input like a task that
# prompts would, and fails for the task named in FAKE_FAIL.
cat >"$work/bin/task" <<'EOF'
#!/usr/bin/env bash
( IFS='|'; echo "$*" ) >>"$FAKE_TASKS"
cat >/dev/null
[ "$1" != "${FAKE_FAIL:-}" ]
EOF
chmod +x "$work/bin/go" "$work/bin/task"

repo="$work/repo"
git init -q "$repo"
git -C "$repo" config user.email test@example.com
git -C "$repo" config user.name Test
mkdir -p "$repo/internal/kernel" "$repo/sdk/go"
echo "module example.com/sdk" >"$repo/sdk/go/go.mod"
echo 'package kernel' >"$repo/internal/kernel/money.go"
git -C "$repo" add . && git -C "$repo" commit -qm init
echo '// changed' >>"$repo/internal/kernel/money.go"
git -C "$repo" add internal/kernel/money.go

# hook runs check-commit.sh in the repository; extra environment comes first in "$@".
hook() {
  : >"$work/log"
  : >"$work/tasks"
  rm -rf "$work/seen"
  (cd "$repo" && env PATH="$work/bin:$PATH" FAKE_LOG="$work/log" FAKE_TASKS="$work/tasks" FAKE_SEEN="$work/seen" "$@" \
    bash "$root/scripts/check-commit.sh" task '. sdk/go' "${FULL:-}") >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" "$work/log" "$work/tasks" >&2
    failures=$((failures + 1))
  fi
}

# same prints yes when the file holds exactly the expected text.
same() {
  if [ "$(cat "$1")" = "$2" ]; then echo yes; else echo no; fi
}

hook FAKE_PLAN='fmt:check:files\tMODULE=.\tFILES=internal/kernel/money.go\nlint:go\tMODULE=.\tPKGS=./a ./b\nkibi:staged\n' && rc=0 || rc=$?
check "a plan runs" "$([ "$rc" -eq 0 ] && echo yes || echo no)"
check "each step runs in order, a variable with spaces stays one argument, stdin doesn't eat the plan" \
  "$(same "$work/tasks" "$(printf 'fmt:check:files|MODULE=.|FILES=internal/kernel/money.go\nlint:go|MODULE=.|PKGS=./a ./b\nkibi:staged')")"
check "the tool reads the staged changes" "$(if tr '\0' '\n' <"$work/seen/changes" | grep -qx internal/kernel/money.go; then echo yes; else echo no; fi)"
check "the tool reads both modules' package listings" \
  "$(if [ -f "$work/seen/go-list/0-tagged.json" ] && [ -f "$work/seen/go-list/1-untagged.json" ]; then echo yes; else echo no; fi)"

hook FAKE_PLAN='lint:go\tMODULE=.\nkibi:staged\n' FAKE_FAIL=lint:go && rc=0 || rc=$?
check "a failing step fails the hook" "$([ "$rc" -ne 0 ] && echo yes || echo no)"
check "a failing step stops the steps after it" "$(same "$work/tasks" 'lint:go|MODULE=.')"

hook FAKE_PLAN='kibi:staged\n' FAKE_TOOL_FAILS=1 && rc=0 || rc=$?
check "a failing selector fails the hook and runs nothing" "$([ "$rc" -ne 0 ] && [ ! -s "$work/tasks" ] && echo yes || echo no)"

FULL=1 hook FAKE_PLAN='check:commit:full\n' && rc=0 || rc=$?
check "FULL selects the full check without listing anything" \
  "$([ "$rc" -eq 0 ] && grep -q -- '-full FULL is set' "$work/log" && ! grep -q 'go list' "$work/log" && echo yes || echo no)"
check "the full check runs" "$(same "$work/tasks" 'check:commit:full')"

git -C "$repo" rev-parse HEAD >"$repo/.git/MERGE_HEAD"
hook FAKE_PLAN='check:commit:full\n' && rc=0 || rc=$?
check "a merge being committed selects the full check" "$([ "$rc" -eq 0 ] && grep -q -- '-full a merge is being committed' "$work/log" && echo yes || echo no)"
rm -f "$repo/.git/MERGE_HEAD"

DRY_RUN=1 hook FAKE_PLAN='kibi:staged\n' && rc=0 || rc=$?
check "DRY_RUN plans and runs nothing" "$([ "$rc" -eq 0 ] && [ ! -s "$work/tasks" ] && grep -q 'go run' "$work/log" && echo yes || echo no)"

if [ "$failures" -gt 0 ]; then
  echo "check-commit.test.sh: $failures of $cases checks failed" >&2
  exit 1
fi
echo "check-commit.test.sh: all $cases checks passed"
