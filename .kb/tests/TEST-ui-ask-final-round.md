---
title: The final round and the retry keep the whole conversation
status: active
sourceFile: internal/views/ui/askfinal_test.go
verification_scope: unit
origin:
  kind: agent
  ref: Live bench finding on R1 and R2 (gpt-oss-120b, two model calls answering with a greeting), 2026-11-14
  recorded_at: '2026-11-14T21:35:06.628Z'
id: TEST-ui-ask-final-round
type: test
---
Runs the loop with a scripted model that answers in plain text first, after tool calls, or with an invalid answer, and checks that the next call holds the system prompt, the question, every tool call and result, the model's reply as an assistant message and then the instruction or the reason, each with its data tier.