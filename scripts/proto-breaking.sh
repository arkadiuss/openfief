#!/usr/bin/env bash
# buf breaking: the protobuf contracts (buf.yaml, proto/) must not break against a base revision.
# PROTO_BASE sets the base (default DOCS_BASE, which CI sets to the merge base, else HEAD).
# A base without contracts has nothing to break, so the check passes there.
#
# A breaking change goes in only when an author decision names it (docs/dev/pipeline.md#breaking-contract-changes):
# scripts/proto-breaking-exceptions.txt lists it, and tools/protobreak excuses exactly the entries new since
# the base, nothing else, and fails on an entry that excuses nothing. Every commit in the range that adds an
# entry must be marked breaking.
#
#   proto-breaking.sh                       check the contracts against the base
#   proto-breaking.sh commit-msg FILE       the commit-msg hook: a staged new entry needs a breaking FILE
set -euo pipefail

exceptions=scripts/proto-breaking-exceptions.txt
tools="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# protobreak builds tools/protobreak from this script's repository, so a test may run the script elsewhere.
protobreak() {
  if [ ! -x "$work/protobreak" ]; then
    (cd "$tools" && go build -o "$work/protobreak" ./tools/protobreak)
  fi
  "$work/protobreak" "$@"
}

# show writes the exceptions file at revision $1 to $2, or nothing when it has none.
show() {
  git show "$1:$exceptions" >"$2" 2>/dev/null || : >"$2"
}

if [ "${1:-}" = commit-msg ]; then
  # A merge brings commits that were checked on their own; the range check below sees each of them.
  if git rev-parse -q --verify MERGE_HEAD >/dev/null || git diff --cached --quiet -- "$exceptions"; then
    exit 0
  fi
  show HEAD "$work/before"
  show "" "$work/after"
  protobreak marked -before "$work/before" -after "$work/after" -message "$2"
  exit 0
fi

base="${PROTO_BASE:-${DOCS_BASE:-HEAD}}"
if [ -z "$(git ls-tree -r --name-only "$base" -- proto 2>/dev/null | grep '\.proto$')" ] ||
  ! git cat-file -e "${base}:buf.yaml" 2>/dev/null; then
  echo "proto:breaking: no contracts at ${base}, nothing to compare"
  exit 0
fi

# buf.lock pins the base's dependencies (such as googleapis), which its contracts import.
mkdir "$work/base"
files=(buf.yaml proto)
if git cat-file -e "${base}:buf.lock" 2>/dev/null; then
  files+=(buf.lock)
fi
git archive "$base" "${files[@]}" | tar -x -C "$work/base"
if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:breaking: every contract at ${base} was removed" >&2
  exit 1
fi

# buf exits 100 when it found breaking changes; anything else but 0 is a failure of its own.
status=0
buf breaking --against "$work/base" --error-format json >"$work/findings" || status=$?
if [ "$status" -ne 0 ] && [ "$status" -ne 100 ]; then
  cat "$work/findings" >&2
  exit "$status"
fi
show "$base" "$work/base-exceptions"
cat "$exceptions" >"$work/exceptions" 2>/dev/null || : >"$work/exceptions"
if [ "$status" -eq 0 ] && cmp -s "$work/base-exceptions" "$work/exceptions"; then
  echo "proto:breaking: no breaking changes against ${base}"
  exit 0
fi

buf build "$work/base" --exclude-imports -o "$work/base.binpb"
protobreak judge -findings "$work/findings" -exceptions "$work/exceptions" -base-exceptions "$work/base-exceptions" \
  -design docs/DESIGN.md -base-image "$work/base.binpb"
for c in $(git rev-list --no-merges "${base}..HEAD" -- "$exceptions"); do
  show "${c}^" "$work/before"
  show "$c" "$work/after"
  git log -1 --format=%B "$c" >"$work/message"
  protobreak marked -before "$work/before" -after "$work/after" -message "$work/message" ||
    { echo "proto:breaking: commit $c" >&2; exit 1; }
done
if git diff --quiet HEAD -- "$exceptions" 2>/dev/null; then
  echo "proto:breaking: only the exceptions above, against ${base}"
else
  echo "proto:breaking: only the exceptions above, against ${base}; the commit that adds them must be marked breaking (the commit-msg hook checks it)"
fi
