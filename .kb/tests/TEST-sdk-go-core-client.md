---
title: The SDK's core client sends the current plugin token as a bearer token
status: active
sourceFile: sdk/go/openfief/core_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:19.182Z'
id: TEST-sdk-go-core-client
type: test
verification_scope: integration
---
An httptest server records each request's Authorization header. CoreHTTPClient with a synthetic token file sends two requests with the file rewritten in between and must send Bearer with the first token and then the rotated one. With no token file the request fails and the server sees nothing.