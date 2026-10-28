#!/usr/bin/env bash
# Tests for scripts/dco-check.sh against a throwaway repository (kibi: TEST-tooling-dco).
set -euo pipefail
# A git hook exports these for the outer repository; the throwaway repository below must not inherit them.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
git init -q
export GIT_AUTHOR_NAME=Alice GIT_AUTHOR_EMAIL=alice@example.invalid
export GIT_COMMITTER_NAME=Alice GIT_COMMITTER_EMAIL=alice@example.invalid
git commit -q --allow-empty -m "chore: base"
base="$(git rev-parse HEAD)"
failures=0

check() { bash "$root/scripts/dco-check.sh" "$base" HEAD 2>&1; }

git commit -q --allow-empty -s -m "feat: signed by the author"
if ! out="$(check)"; then
  echo "FAIL: a signed-off commit should pass, got: $out" >&2
  failures=$((failures + 1))
fi

git commit -q --allow-empty -m "fix: signed by someone else" -m "Signed-off-by: Bob <bob@example.invalid>"
if out="$(check)"; then
  echo "FAIL: a sign-off by someone other than the author should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'fix: signed by someone else' <<<"$out"; then
  echo "FAIL: the failure should name the commit, got: $out" >&2
  failures=$((failures + 1))
fi
git reset -q --hard HEAD~1

git commit -q --allow-empty -m "docs: not signed"
if out="$(check)"; then
  echo "FAIL: a commit without a sign-off should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'docs: not signed' <<<"$out"; then
  echo "FAIL: the failure should name the commit, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "dco-check: 3 cases passed"
