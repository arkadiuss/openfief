#!/usr/bin/env bash
# Tests for scripts/check-docs.sh against a throwaway copy of the docs (kibi: TEST-tooling-docs-check).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/docs" "$work/scripts"
cp "$root/docs/DESIGN.md" "$root/docs/SPEC.md" "$work/docs/"
cp "$root/scripts/check-docs.sh" "$work/scripts/"
git -C "$work" init -q
git -C "$work" -c user.name=test -c user.email=test@example.invalid add -A
git -C "$work" -c user.name=test -c user.email=test@example.invalid commit -q -m baseline

cd "$work"
failures=0

if ! bash scripts/check-docs.sh >/dev/null 2>&1; then
  echo "FAIL: unchanged docs should pass" >&2
  failures=$((failures + 1))
fi

cp docs/DESIGN.md design.bak
printf '\nSee OQ-999.\n' >>docs/DESIGN.md
if out="$(bash scripts/check-docs.sh 2>&1)"; then
  echo "FAIL: an undefined OQ tag should fail the check" >&2
  failures=$((failures + 1))
elif ! grep -q 'undefined OQ tags: OQ-999' <<<"$out"; then
  echo "FAIL: the undefined tag should be named, got: $out" >&2
  failures=$((failures + 1))
fi
cp design.bak docs/DESIGN.md

sed -i 's/^- \*\*\[GR-14\]\*\* \(.*\)$/- **[GR-14]** \1 (edited)/' docs/DESIGN.md
if out="$(bash scripts/check-docs.sh 2>&1)"; then
  echo "FAIL: an edited ground rule should fail the check" >&2
  failures=$((failures + 1))
elif ! grep -q 'ground-rule lines' <<<"$out"; then
  echo "FAIL: the ground-rule change should be named, got: $out" >&2
  failures=$((failures + 1))
fi
cp design.bak docs/DESIGN.md

if [ "$failures" -ne 0 ]; then
  exit 1
fi
echo "check-docs: 3 cases passed"
