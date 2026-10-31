---
title: proto-breaking.sh passes compatible contract changes and fails breaking ones against the base
status: active
sourceFile: scripts/tests/proto-breaking.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T07:37:11.504Z'
id: TEST-tooling-proto-breaking
type: test
verification_scope: integration
---
Builds a throwaway git repository with the project's buf.yaml. With no contracts at the base revision the script passes and says so. After committing a contract, adding a field passes, while changing a field's type fails and names the field. After committing the breaking change, HEAD against itself passes, and the same commit fails when PROTO_BASE or DOCS_BASE points at the previous revision. Runs the pinned buf from the task test target.