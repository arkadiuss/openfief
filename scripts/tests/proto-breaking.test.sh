#!/usr/bin/env bash
# Tests for scripts/proto-breaking.sh against a throwaway repository (kibi: TEST-tooling-proto-breaking).
# Needs buf on PATH (task test puts the pinned one there).
set -euo pipefail
# A git hook exports these for the outer repository; the throwaway repository below must not inherit them.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE PROTO_BASE DOCS_BASE

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"
git init -q
export GIT_AUTHOR_NAME=Alice GIT_AUTHOR_EMAIL=alice@example.invalid
export GIT_COMMITTER_NAME=Alice GIT_COMMITTER_EMAIL=alice@example.invalid
failures=0

check() { bash "$root/scripts/proto-breaking.sh" 2>&1; }

git commit -q --allow-empty -m "chore: no contracts yet"
mkdir -p proto/test/v1
cp "$root/buf.yaml" buf.yaml
cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

message Thing {
  string name = 1;
}
PROTO

if ! out="$(check)"; then
  echo "FAIL: a base without contracts should pass, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'no contracts at HEAD' <<<"$out"; then
  echo "FAIL: a base without contracts should say so, got: $out" >&2
  failures=$((failures + 1))
fi

git add -A && git commit -q -m "feat: add contracts"
cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

message Thing {
  string name = 1;
  string note = 2;
}
PROTO
if ! out="$(check)"; then
  echo "FAIL: adding a field should pass, got: $out" >&2
  failures=$((failures + 1))
fi

cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

message Thing {
  int64 name = 1;
}
PROTO
if out="$(check)"; then
  echo "FAIL: changing a field's type should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'name' <<<"$out"; then
  echo "FAIL: the failure should name the field, got: $out" >&2
  failures=$((failures + 1))
fi

base="$(git rev-parse HEAD)"
git add -A && git commit -q -m "feat!: break the contract"
if ! out="$(check)"; then
  echo "FAIL: HEAD against itself should pass, got: $out" >&2
  failures=$((failures + 1))
fi
if out="$(PROTO_BASE="$base" check)"; then
  echo "FAIL: a breaking commit against PROTO_BASE should fail" >&2
  failures=$((failures + 1))
fi
if out="$(DOCS_BASE="$base" check)"; then
  echo "FAIL: a breaking commit against DOCS_BASE should fail" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  echo "proto-breaking.test.sh: $failures failure(s)" >&2
  exit 1
fi
echo "proto-breaking.test.sh: ok"
