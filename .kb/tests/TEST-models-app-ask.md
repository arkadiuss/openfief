---
title: The Ask box's door checks the session, answers availability and tier questions without a call or an entry, and calls the router for the ask role
status: active
sourceFile: internal/modelrouter/app/ask_test.go
verification_scope: unit
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:30.136Z'
id: TEST-models-app-ask
type: test
---
Table tests over the router with fakes: the switch, toggle and route, valid and unknown sessions, tiers that may and may not go, and the audit entries of a call.