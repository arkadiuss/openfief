#!/usr/bin/env bash
# buf breaking: the protobuf contracts (buf.yaml, proto/) must not break against a base revision.
# PROTO_BASE sets the base (default DOCS_BASE, which CI sets to the merge base, else HEAD).
# A base without contracts has nothing to break, so the check passes there.
set -euo pipefail

base="${PROTO_BASE:-${DOCS_BASE:-HEAD}}"
if [ -z "$(git ls-tree -r --name-only "$base" -- proto 2>/dev/null | grep '\.proto$')" ] ||
  ! git cat-file -e "${base}:buf.yaml" 2>/dev/null; then
  echo "proto:breaking: no contracts at ${base}, nothing to compare"
  exit 0
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# buf.lock pins the base's dependencies (such as googleapis), which its contracts import.
files=(buf.yaml proto)
if git cat-file -e "${base}:buf.lock" 2>/dev/null; then
  files+=(buf.lock)
fi
git archive "$base" "${files[@]}" | tar -x -C "$work"
if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:breaking: every contract at ${base} was removed" >&2
  exit 1
fi
buf breaking --against "$work"
echo "proto:breaking: no breaking changes against ${base}"
