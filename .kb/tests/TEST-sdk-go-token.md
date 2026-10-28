---
title: TokenFile reads the projected token from its file on every call
status: active
sourceFile: sdk/go/openfief/token_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:24.730Z'
id: TEST-sdk-go-token
type: test
---
Table test over a temporary token file: a token is returned trimmed of surrounding whitespace, an empty or blank file fails with ErrEmptyToken, and a missing file fails with fs.ErrNotExist. A second test rewrites the file between calls, as the kubelet does when it rotates a projected token, and expects each call to return the token current at that moment, which rules out caching.