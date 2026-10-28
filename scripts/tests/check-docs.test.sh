#!/usr/bin/env bash
# Tests for scripts/check-docs.sh against a throwaway copy of the docs (kibi: TEST-tooling-docs-check).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/docs" "$work/scripts"
cp -R "$root/docs/DESIGN.md" "$root/docs/glossary.md" "$root/docs/research.md" "$root/docs/proposals" "$work/docs/"
cp "$root/scripts/check-docs.sh" "$work/scripts/"
git -C "$work" init -q
git -C "$work" -c user.name=test -c user.email=test@example.invalid add -A
git -C "$work" -c user.name=test -c user.email=test@example.invalid commit -q -m baseline

cd "$work"
failures=0

expect_fail() {
  local name="$1" pattern="$2" out
  if out="$(bash scripts/check-docs.sh 2>&1)"; then
    echo "FAIL: $name should fail the check" >&2
    failures=$((failures + 1))
  elif ! grep -q "$pattern" <<<"$out"; then
    echo "FAIL: $name should report '$pattern', got: $out" >&2
    failures=$((failures + 1))
  fi
  git checkout -q -- docs
}

if ! bash scripts/check-docs.sh >/dev/null 2>&1; then
  echo "FAIL: unchanged docs should pass" >&2
  failures=$((failures + 1))
fi

printf '\nSee OQ-999.\n' >>docs/DESIGN.md
expect_fail "an undefined OQ tag" 'undefined OQ tags: OQ-999'

sed -i 's/^- \*\*\[GR-14\]\*\* \(.*\)$/- **[GR-14]** \1 (edited)/' docs/DESIGN.md
expect_fail "an edited ground rule" 'ground-rule lines'

printf '\n- **[OQ-21]** A second definition.\n' >>docs/proposals/hosting.md
expect_fail "an ID defined twice" 'OQ IDs defined more than once: OQ-21'

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "check-docs: 4 cases passed"
