---
title: The router enforces the switch, the tiers and the audit before it calls
status: active
sourceFile: internal/modelrouter/app/router_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:27:58.874Z'
id: TEST-models-app-router
type: test
---
Runs the router over fake settings, providers and audit log: switch off calls nothing, entries are written before the call and a failed write refuses it, tiers and fallbacks, unavailable and rejected providers, the entry cache, and that records and errors hold no text.