#!/usr/bin/env bash
# Fails when the generated Go code in sdk/go/gen differs from what buf.gen.yaml generates from proto/.
set -euo pipefail

if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:check: no contracts yet, nothing generated"
  exit 0
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
buf generate -o "$work"
if ! diff -r "$work/sdk/go/gen" sdk/go/gen; then
  echo "proto:check: sdk/go/gen is stale; run task proto:gen and commit the result" >&2
  exit 1
fi
echo "proto:check: generated code is up to date"
