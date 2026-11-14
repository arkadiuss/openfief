---
title: AdmitAskQuestion and GenerateAsk answer the guards by kind and carry the reset
status: active
sourceFile: internal/modelrouter/connectapi/ask_guards_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:53.201Z'
id: TEST-models-connectapi-ask-guards
type: test
---
Through the generated Connect client: each admission decision with its reset time, nothing counted when the budget is spent, the session invalid as unauthenticated, a budget refusal of GenerateAsk as RESOURCE_EXHAUSTED with the refusal and reset metadata and one audit entry of class budget, and a busy refusal while a call holds the only place.