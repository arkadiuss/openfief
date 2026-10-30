---
title: The selector reads its inputs and prints its decisions
status: active
sourceFile: tools/affected/load_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:27.154Z'
id: TEST-tooling-affected-inputs
type: test
---
Tests parse git's NUL-separated name-status output, the $root paths of script tests, the builds file and the image matrix, go list's JSON (including a package outside the repository, which is refused) and the relative path literals a package's Go files name. End-to-end through run() on an in-memory repository: hook prints the steps tab-separated and the reasons, and the full check for go.mod and -full; ci prints code=false only for a docs-only pull request and code=true for a push; images prints the matrix to build and the images to carry for reuse, base and all modes; bad flags and missing inputs fail.