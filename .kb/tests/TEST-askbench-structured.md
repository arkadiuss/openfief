---
title: The bench's structured-answer path, schema failures and kind-decided cases against a fake model
status: active
sourceFile: tools/askbench/structured_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:56.195Z'
id: TEST-askbench-structured
type: test
---
Runs the command against a fake OpenAI-compatible server: a model with the structured-output flag is asked the answer as a strict json_schema and one without as json_object, only in rounds that offer no tools, the flag is in the configuration hash, a reply that is no structured answer after its retry is a schema failure with its reason, and R1, R2, R5 and A18 pass by the answer's kind, suggestions and candidates.
