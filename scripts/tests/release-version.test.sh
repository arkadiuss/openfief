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
failures=0

expect() {
  local want="$1" channel="$2" version="$3" got=pass
  bash "$root/scripts/release-version.sh" "$channel" "$version" >/dev/null 2>&1 || got=fail
  if [ "$got" != "$want" ]; then
    echo "FAIL: $channel $version should $want, got $got" >&2
    failures=$((failures + 1))
  fi
}

cases=0
while read -r want channel version; do
  expect "$want" "$channel" "$version"
  cases=$((cases + 1))
done <<'EOF'
pass stable 2026.11.0
pass stable 2026.1.12
pass beta 2026.11.0-beta.1
pass beta 2026.12.3-beta.10
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
EOF

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-version: $cases cases passed"
