---
title: The core wires the router's audit entries and providers
status: active
sourceFile: cmd/core/modelrouter_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:06.456Z'
id: TEST-core-model-router
type: test
---
Checks that a call becomes a valid model_call audit entry, that a failing log fails the call, and that providers are unavailable before discovery.