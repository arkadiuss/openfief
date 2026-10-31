---
title: The built-in UI serves only requests that carry the core's token, and probes
status: active
sourceFile: plugins/ui/core_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:48.366Z'
id: TEST-ui-require-core
type: test
verification_scope: unit
---
Wraps a handler with the UI's requireCore and a fake core verifier: a request with the core's token reaches the page; none, or a session token in Authorization, gets 401 before any page runs; GET /healthz answers without a token for probes, and POST /healthz does not.