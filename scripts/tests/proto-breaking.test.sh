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

# An author decision's removal (D-157's shape): a method and its messages, excused by the entries that
# go in with it, in a commit marked breaking, and nothing else.
mkdir -p docs scripts
cat >docs/DESIGN.md <<'MD'
- **[D-9]** The `ListGadgets` method and its messages are removed from the contract.
- **[D-10]** Things are listed by name.
MD
cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

service Things {
  rpc GetThing(GetThingRequest) returns (Thing);
  rpc ListGadgets(ListGadgetsRequest) returns (ListGadgetsResponse);
}

message Thing {
  int64 name = 1;
}

message GetThingRequest {}

message ListGadgetsRequest {}

message ListGadgetsResponse {
  repeated Gadget gadgets = 1;
}

message Gadget {
  string id = 1;
}
PROTO
git add -A && git commit -q -m "feat: things and gadgets"
before_removal="$(git rev-parse HEAD)"
cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

service Things {
  rpc GetThing(GetThingRequest) returns (Thing);
}

message Thing {
  int64 name = 1;
}

message GetThingRequest {}
PROTO
if out="$(check)"; then
  echo "FAIL: removing a method without an exception should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'not excused.*ListGadgets' <<<"$out"; then
  echo "FAIL: the failure should name the removed method, got: $out" >&2
  failures=$((failures + 1))
fi
cat >scripts/proto-breaking-exceptions.txt <<'TXT'
# synthetic
D-9 RPC_NO_DELETE test.v1.Things.ListGadgets
D-9 MESSAGE_NO_DELETE test.v1.ListGadgetsRequest
D-9 MESSAGE_NO_DELETE test.v1.ListGadgetsResponse
D-9 MESSAGE_NO_DELETE test.v1.Gadget
TXT
if ! out="$(check)"; then
  echo "FAIL: a removal its decision names should pass with its entries, got: $out" >&2
  failures=$((failures + 1))
elif ! grep -q 'D-9 excuses RPC_NO_DELETE test.v1.Things.ListGadgets' <<<"$out"; then
  echo "FAIL: the check should say what it excused, got: $out" >&2
  failures=$((failures + 1))
fi

# Any other breaking change in the same change still fails.
sed -i 's/int64 name = 1;/bool name = 1;/' proto/test/v1/test.proto
if out="$(check)"; then
  echo "FAIL: another breaking change next to the excused removal should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'not excused.*name' <<<"$out"; then
  echo "FAIL: the failure should name the other change, got: $out" >&2
  failures=$((failures + 1))
fi
sed -i 's/bool name = 1;/int64 name = 1;/' proto/test/v1/test.proto

# An entry citing a decision that doesn't name the method fails.
cp scripts/proto-breaking-exceptions.txt "$work/entries"
sed -i 's/^D-9 RPC_NO_DELETE/D-10 RPC_NO_DELETE/' scripts/proto-breaking-exceptions.txt
if out="$(check)"; then
  echo "FAIL: an entry whose decision doesn't name the method should fail" >&2
  failures=$((failures + 1))
elif ! grep -q "D-10 doesn't name" <<<"$out"; then
  echo "FAIL: the failure should say the decision doesn't name it, got: $out" >&2
  failures=$((failures + 1))
fi
cp "$work/entries" scripts/proto-breaking-exceptions.txt

# The commit-msg hook: a staged new entry needs a commit marked breaking.
git add -A
printf 'feat(views): remove ListGadgets (D-9)\n' >"$work/msg"
if out="$(bash "$root/scripts/proto-breaking.sh" commit-msg "$work/msg" 2>&1)"; then
  echo "FAIL: an unmarked commit adding an entry should fail at commit-msg" >&2
  failures=$((failures + 1))
fi
printf 'feat(views)!: remove ListGadgets (D-9)\n' >"$work/msg"
if ! out="$(bash "$root/scripts/proto-breaking.sh" commit-msg "$work/msg" 2>&1)"; then
  echo "FAIL: a breaking commit adding an entry should pass at commit-msg, got: $out" >&2
  failures=$((failures + 1))
fi

# In CI, against the base, every commit adding an entry must be marked breaking.
git commit -q -m "feat(views): remove ListGadgets (D-9)"
if out="$(PROTO_BASE="$before_removal" check)"; then
  echo "FAIL: an unmarked commit adding an entry should fail against the base" >&2
  failures=$((failures + 1))
elif ! grep -q 'must be marked breaking' <<<"$out"; then
  echo "FAIL: the failure should ask for the breaking mark, got: $out" >&2
  failures=$((failures + 1))
fi
git commit -q --amend -m "feat(views)!: remove ListGadgets (D-9)"
if ! out="$(PROTO_BASE="$before_removal" check)"; then
  echo "FAIL: a breaking commit with its removal and entries should pass against the base, got: $out" >&2
  failures=$((failures + 1))
fi

# Once the entries are on the base they excuse nothing again: a later removal fails, and so does an
# entry added for nothing this change breaks.
if ! out="$(check)"; then
  echo "FAIL: the committed removal against itself should pass, got: $out" >&2
  failures=$((failures + 1))
fi
cat >proto/test/v1/test.proto <<'PROTO'
syntax = "proto3";

package test.v1;

service Things {}

message Thing {
  int64 name = 1;
}

message GetThingRequest {}
PROTO
if out="$(check)"; then
  echo "FAIL: an old entry should not excuse a later removal" >&2
  failures=$((failures + 1))
elif ! grep -q 'not excused.*GetThing' <<<"$out"; then
  echo "FAIL: the failure should name the later removal, got: $out" >&2
  failures=$((failures + 1))
fi
git checkout -q -- proto
echo "D-9 RPC_NO_DELETE test.v1.Things.GetThing" >>scripts/proto-breaking-exceptions.txt
if out="$(check)"; then
  echo "FAIL: an entry that excuses nothing should fail" >&2
  failures=$((failures + 1))
elif ! grep -q 'excuses nothing' <<<"$out"; then
  echo "FAIL: the failure should say the entry excuses nothing, got: $out" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  echo "proto-breaking.test.sh: $failures failure(s)" >&2
  exit 1
fi
echo "proto-breaking.test.sh: ok"
