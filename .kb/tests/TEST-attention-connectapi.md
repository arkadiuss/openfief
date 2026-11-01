---
title: The AttentionService admits only the owner's view and maps its errors
status: active
sourceFile: internal/attention/connectapi/attention_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:42.192Z'
id: TEST-attention-connectapi
type: test
---
Serves the AttentionService on an httptest server over attention/app.Items with an in-memory store and calls it with the generated client: the owner acknowledges and resolves; a read session, no session and a caller that is not a view plugin are refused with their codes; an unknown item is not found, a kind that resolves on its own fails its precondition, and a store failure is an internal error that leaks nothing.