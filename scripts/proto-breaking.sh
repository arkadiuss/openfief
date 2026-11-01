#!/usr/bin/env bash
# buf breaking: the protobuf contracts (buf.yaml, proto/) must not break against a base revision.
# PROTO_BASE sets the base (default DOCS_BASE, which CI sets to the merge base, else HEAD).
# A base without contracts has nothing to break, so the check passes there.
#
# What may break (docs/dev/pipeline.md#breaking-contract-changes):
# - Until the first beta (no beta or stable release tag yet, tools/protobreak stage), any breaking change, in a
#   commit marked breaking (`!` or a BREAKING CHANGE: footer): every commit in the range that breaks a
#   contract against its parent must be marked, and so must the commit being written (commit-msg).
# - From the first beta, only what an author decision names: scripts/proto-breaking-exceptions.txt lists it,
#   and tools/protobreak excuses exactly the entries new since the base, nothing else.
# Either way an exceptions entry must cite a decision that names it and excuse something this change breaks,
# and every commit that adds one must be marked breaking.
#
#   proto-breaking.sh                       check the contracts against the base
#   proto-breaking.sh commit-msg FILE       the commit-msg hook: a staged break or new entry needs a breaking FILE
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

# contracts extracts the contracts of revision $1 into the new directory $2, with the revision's own
# buf.lock, which pins its dependencies (such as googleapis); it fails when the revision has none.
contracts() {
  if [ -z "$(git ls-tree -r --name-only "$1" -- proto 2>/dev/null | grep '\.proto$')" ] ||
    ! git cat-file -e "$1:buf.yaml" 2>/dev/null; then
    return 1
  fi
  mkdir -p "$2"
  local files=(buf.yaml proto)
  if git cat-file -e "$1:buf.lock" 2>/dev/null; then
    files+=(buf.lock)
  fi
  git archive "$1" "${files[@]}" | tar -x -C "$2"
}

# breaks runs buf breaking on the contracts in $1 against those in $2, writing its JSON findings to $3. It
# succeeds when they break, fails when they don't, and stops the script when buf fails on its own.
breaks() {
  local status=0
  buf breaking "$1" --against "$2" --error-format json >"$3" || status=$?
  case "$status" in
    0) return 1 ;;
    100) return 0 ;;
    *) cat "$3" >&2; exit "$status" ;;
  esac
}

# stage prints pre-beta or beta from the repository's tags.
stage() {
  git tag -l >"$work/tags"
  protobreak stage <"$work/tags"
}

if [ "${1:-}" = commit-msg ]; then
  # A merge brings commits that were checked on their own; the range check below sees each of them.
  if git rev-parse -q --verify MERGE_HEAD >/dev/null; then
    exit 0
  fi
  if ! git diff --cached --quiet -- "$exceptions"; then
    show HEAD "$work/before"
    show "" "$work/after"
    protobreak marked -before "$work/before" -after "$work/after" -message "$2"
  fi
  if ! git diff --cached --quiet -- proto buf.yaml && [ "$(stage)" = pre-beta ] &&
    contracts HEAD "$work/head" && breaks . "$work/head" "$work/findings"; then
    protobreak marked -breaking -message "$2"
  fi
  exit 0
fi

base="${PROTO_BASE:-${DOCS_BASE:-HEAD}}"
if ! contracts "$base" "$work/base"; then
  echo "proto:breaking: no contracts at ${base}, nothing to compare"
  exit 0
fi
if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:breaking: every contract at ${base} was removed" >&2
  exit 1
fi

broken=false
if breaks . "$work/base" "$work/findings"; then
  broken=true
fi
show "$base" "$work/base-exceptions"
cat "$exceptions" >"$work/exceptions" 2>/dev/null || : >"$work/exceptions"
if ! $broken && cmp -s "$work/base-exceptions" "$work/exceptions"; then
  echo "proto:breaking: no breaking changes against ${base}"
  exit 0
fi

buf build "$work/base" --exclude-imports -o "$work/base.binpb"
git tag -l >"$work/tags"
protobreak judge -findings "$work/findings" -exceptions "$work/exceptions" -base-exceptions "$work/base-exceptions" \
  -design docs/DESIGN.md -base-image "$work/base.binpb" -tags "$work/tags"
for c in $(git rev-list --no-merges "${base}..HEAD" -- "$exceptions"); do
  show "${c}^" "$work/before"
  show "$c" "$work/after"
  git log -1 --format=%B "$c" >"$work/message"
  protobreak marked -before "$work/before" -after "$work/after" -message "$work/message" ||
    { echo "proto:breaking: commit $c" >&2; exit 1; }
done
# Before the first beta, each commit that breaks a contract against its parent must be marked breaking.
if $broken && [ "$(stage)" = pre-beta ]; then
  for c in $(git rev-list --no-merges "${base}..HEAD" -- proto buf.yaml); do
    rm -rf "$work/c" "$work/p"
    if contracts "$c" "$work/c" && contracts "${c}^" "$work/p" && breaks "$work/c" "$work/p" "$work/commit-findings"; then
      git log -1 --format=%B "$c" >"$work/message"
      protobreak marked -breaking -message "$work/message" || { echo "proto:breaking: commit $c" >&2; exit 1; }
    fi
  done
fi
if git diff --quiet HEAD -- "$exceptions" proto buf.yaml 2>/dev/null; then
  echo "proto:breaking: only the changes above, against ${base}"
else
  echo "proto:breaking: only the changes above, against ${base}; the commit that makes them must be marked breaking (the commit-msg hook checks it)"
fi
