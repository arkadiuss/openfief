---
title: A round traces the reasoning length and no reasoning text
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:14.936Z'
sourceFile: tools/askbench/finalround_test.go
verification_scope: integration
id: TEST-askbench-reasoning-length-trace
type: test
---
Runs the same fake server with a reasoning field on the empty reply and checks that the round's reasoningChars is its length and that report.json holds no reasoning text.

## Context

Trace check of the reasoning length.