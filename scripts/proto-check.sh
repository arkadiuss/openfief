#!/usr/bin/env bash
# Fails when the generated code in sdk/go/gen or gen/openapi differs from what buf.gen.yaml generates from proto/.
set -euo pipefail

if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:check: no contracts yet, nothing generated"
  exit 0
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
buf generate -o "$work"
for dir in sdk/go/gen gen/openapi; do
  if ! diff -r "$work/$dir" "$dir"; then
    echo "proto:check: $dir is stale; run task proto:gen and commit the result" >&2
    exit 1
  fi
done
echo "proto:check: generated code is up to date"
