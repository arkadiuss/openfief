---
title: The front door logs a path without its query
status: active
sourceFile: internal/views/browserproxy/logs_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T08:17:43.562Z'
id: TEST-browserproxy-logs
type: test
---
Drives the proxy with a request carrying a synthetic question in its query, toward several views and toward a view that doesn't answer, and checks the log names the path and never the question.