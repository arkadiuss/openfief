#!/usr/bin/env bash
# Tests for scripts/release-notes.sh against a throwaway repository (kibi: TEST-tooling-release-notes).
set -euo pipefail
# A git hook exports these for the outer repository; the throwaway repository below must not inherit them.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
git init -q
export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.invalid
export GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.invalid
failures=0

commit() { git commit -q --allow-empty "$@"; }

expect_notes() {
  local name="$1" version="$2" want="$3" got
  got="$(bash "$root/scripts/release-notes.sh" "$version" | sed -E 's/ \([0-9a-f]+\)$//')"
  if [ "$got" != "$want" ]; then
    printf 'FAIL: %s\nwant:\n%s\ngot:\n%s\n' "$name" "$want" "$got" >&2
    failures=$((failures + 1))
  fi
}

commit -m "chore: base"
git tag 2026.10.0
commit -m "feat(ledger): add lots"
git tag 2026.11.0-beta.1
commit -m "fix: round half to even"
commit -m "feat(ingest)!: rename the import batch API"
commit -m "refactor: tidy" -m "BREAKING CHANGE: drops the old flag"
commit -m "perf(valuation): cache FX rates"
commit -m "Update README"

expect_notes "stable notes start after the previous stable" 2026.11.0 "## 2026.11.0

Changes since 2026.10.0.

### Breaking changes

- refactor: tidy
- feat(ingest)!: rename the import batch API

### Features

- feat(ledger): add lots

### Fixes

- fix: round half to even

### Performance

- perf(valuation): cache FX rates

### Other changes

- Update README"

expect_notes "beta notes start after the previous release of either channel" 2026.11.0-beta.2 "## 2026.11.0-beta.2

Changes since 2026.11.0-beta.1.

### Breaking changes

- refactor: tidy
- feat(ingest)!: rename the import batch API

### Fixes

- fix: round half to even

### Performance

- perf(valuation): cache FX rates

### Other changes

- Update README"

# An alpha is a pre-release like a beta: its notes start after the newest release of any channel, and a stable
# release skips it too.
git tag 2026.11.0-alpha.1 HEAD~2
expect_notes "alpha notes start after the previous release of any channel" 2026.11.0-alpha.2 "## 2026.11.0-alpha.2

Changes since 2026.11.0-alpha.1.

### Performance

- perf(valuation): cache FX rates

### Other changes

- Update README"

expect_notes "stable notes skip alpha releases" 2026.11.0 "## 2026.11.0

Changes since 2026.10.0.

### Breaking changes

- refactor: tidy
- feat(ingest)!: rename the import batch API

### Features

- feat(ledger): add lots

### Fixes

- fix: round half to even

### Performance

- perf(valuation): cache FX rates

### Other changes

- Update README"

git tag -d 2026.10.0 2026.11.0-beta.1 2026.11.0-alpha.1 >/dev/null
expect_notes "the first release covers the whole history" 2026.11.0 "## 2026.11.0

First release.

### Breaking changes

- refactor: tidy
- feat(ingest)!: rename the import batch API

### Features

- feat(ledger): add lots

### Fixes

- fix: round half to even

### Performance

- perf(valuation): cache FX rates

### Other changes

- Update README
- chore: base"

# Known limits come from the version's own file at the ref, ahead of the changes; another version's file is ignored.
git tag 2026.11.0-alpha.1
mkdir -p .github/release-notes/2026.11.0-alpha.2 .github/release-notes/2026.11.0-alpha.3
printf -- '- The demo has no price checks yet.\n- Bonds are valued at par.\n\n' >.github/release-notes/2026.11.0-alpha.2/known-limits.md
printf -- '- Only for alpha.3.\n' >.github/release-notes/2026.11.0-alpha.3/known-limits.md
git add .github
commit -m "docs: known limits of the next alphas"
expect_notes "known limits from the version's file come first" 2026.11.0-alpha.2 "## 2026.11.0-alpha.2

Changes since 2026.11.0-alpha.1.

### Known limits

- The demo has no price checks yet.
- Bonds are valued at par.

### Other changes

- docs: known limits of the next alphas"

expect_notes "a version without a file has no known limits" 2026.11.0-alpha.4 "## 2026.11.0-alpha.4

Changes since 2026.11.0-alpha.1.

### Other changes

- docs: known limits of the next alphas"

printf '\n  \n' >.github/release-notes/2026.11.0-alpha.3/known-limits.md
git add .github
commit -m "docs: empty the known limits"
if out="$(bash "$root/scripts/release-notes.sh" 2026.11.0-alpha.3 2>&1)"; then
  printf 'FAIL: an empty known-limits file should fail, got:\n%s\n' "$out" >&2
  failures=$((failures + 1))
elif ! grep -q 'known-limits.md is empty' <<<"$out"; then
  printf 'FAIL: an empty known-limits file should be named, got:\n%s\n' "$out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "release-notes: 8 cases passed"
