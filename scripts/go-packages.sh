#!/usr/bin/env bash
# Lists the Go packages of each module for tools/affected (docs/dev/pipeline.md#affected-checks):
# <out>/go-list/<i>-tagged.json with the e2e build tag, for imports and embedded files with the end-to-end suite
# included, and <out>/go-list/<i>-untagged.json without, for what `go test` can run; i is the module's position.
# Usage: go-packages.sh <absolute output dir> <module dir>...
set -euo pipefail

out="$1"
shift
mkdir -p "$out/go-list"
i=0
for module in "$@"; do
  (cd "$module" && go list -e -json -tags e2e ./... >"$out/go-list/$i-tagged.json")
  (cd "$module" && go list -e -json ./... >"$out/go-list/$i-untagged.json")
  i=$((i + 1))
done
