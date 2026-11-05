#!/usr/bin/env bash
# Tests for scripts/release-version.sh against a throwaway repository (kibi: TEST-tooling-release-version).
set -euo pipefail
# A git hook exports these for the outer repository; the throwaway repository below must not inherit them.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
git init -q
git -c user.name=test -c user.email=test@example.invalid commit -q --allow-empty -m "chore: base"
git tag 2026.10.0
git tag 2026.10.0-alpha.1
# The release commit of a publish that created its tag and then failed, which a re-run resumes: a lightweight tag,
# as the GitHub API creates it, and an annotated one. HEAD is a later commit, so HEAD~1 is the release commit.
git tag -a -m "2026.10.0-beta.1" 2026.10.0-beta.1
git -c user.name=test -c user.email=test@example.invalid commit -q --allow-empty -m "chore: later"
failures=0

expect() {
  local want="$1" channel="$2" version="$3" commit="${4:-}" got=pass
  bash "$root/scripts/release-version.sh" "$channel" "$version" ${commit:+"$commit"} >/dev/null 2>&1 || got=fail
  if [ "$got" != "$want" ]; then
    echo "FAIL: $channel $version ${commit:+at $commit }should $want, got $got" >&2
    failures=$((failures + 1))
  fi
}

# The optional fourth column is the release commit the publish job passes.
cases=0
while read -r want channel version commit; do
  expect "$want" "$channel" "$version" "$commit"
  cases=$((cases + 1))
done <<'EOF'
pass stable 2026.11.0
pass stable 2026.1.12
pass beta 2026.11.0-beta.1
pass beta 2026.12.3-beta.10
pass alpha 2026.11.0-alpha.1
pass alpha 2026.12.3-alpha.10
pass alpha 2026.10.0-alpha.2
fail stable v2026.11.0
fail stable 2026.11.0-beta.1
fail beta 2026.11.0
fail beta 2026.11.0-beta.0
fail stable 2026.01.0
fail stable 2026.13.0
fail stable 2026.11.01
fail stable 2026.11
fail nightly 2026.11.0
fail stable 2026.10.0
fail alpha 2026.11.0
fail alpha 2026.11.0-alpha.0
fail alpha 2026.11.0-alpha.01
fail alpha v2026.11.0-alpha.1
fail alpha 2026.11.0-alpha
fail alpha 2026.11.0-ALPHA.1
fail alpha 2026.11.0-beta.1
fail beta 2026.11.0-alpha.1
fail stable 2026.11.0-alpha.1
fail alpha 2026.10.0-alpha.1
fail beta 2026.10.0-beta.1
pass alpha 2026.10.0-alpha.1 HEAD~1
pass beta 2026.10.0-beta.1 HEAD~1
pass alpha 2026.11.0-alpha.1 HEAD
fail alpha 2026.10.0-alpha.1 HEAD
fail beta 2026.10.0-beta.1 HEAD
fail alpha 2026.10.0-alpha.1 no-such-commit
fail alpha 2026.11.0-alpha.0 HEAD
EOF

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-version: $cases cases passed"
