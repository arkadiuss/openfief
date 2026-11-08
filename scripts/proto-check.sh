#!/usr/bin/env bash
# Fails when the generated code in sdk/go/gen or gen/openapi differs from what buf.gen.yaml generates from proto/,
# with tools/openapierrors's error responses added to the read API's OpenAPI document as task proto:gen adds them.
set -euo pipefail

if [ -z "$(find proto -name '*.proto' 2>/dev/null)" ]; then
  echo "proto:check: no contracts yet, nothing generated"
  exit 0
fi

: "${OPENAPI_SPEC:=gen/openapi/openfief/views/v1/read_service.openapi.yaml}"
: "${OPENAPI_ERRORS:=proto/openfief/views/v1/read_service.openapi-errors.yaml}"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
buf generate -o "$work"
go run ./tools/openapierrors -rules "$OPENAPI_ERRORS" -spec "$work/$OPENAPI_SPEC"
for dir in sdk/go/gen gen/openapi; do
  if ! diff -r "$work/$dir" "$dir"; then
    echo "proto:check: $dir is stale; run task proto:gen and commit the result" >&2
    exit 1
  fi
done
echo "proto:check: generated code is up to date"
