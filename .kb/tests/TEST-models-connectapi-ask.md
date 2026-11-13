---
title: The AskModelService translates messages and tiers, maps the router's errors by kind and keeps text out of errors and logs
status: active
sourceFile: internal/modelrouter/connectapi/ask_test.go
verification_scope: integration
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:30.863Z'
id: TEST-models-connectapi-ask
type: test
---
Calls the Connect handler through a client over a real router with a fake provider and audit log.