---
title: The bench grades an empty reply after a tool call followed by the answer instruction as a pass
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:14.783Z'
sourceFile: tools/askbench/finalround_test.go
verification_scope: integration
id: TEST-askbench-final-round-empty-reply
type: test
---
Runs the command against a fake OpenAI-compatible server that calls the value tool, answers empty after its result, and answers from the card only when the final round asks for the answer, and checks that A1 passes after three calls and that no request restates a reply.

## Context

Bench check of the empty-reply path.