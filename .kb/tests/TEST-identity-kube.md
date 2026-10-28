---
title: The core's tokens to plugin instances are requested per instance and cached until shortly before expiry
status: active
sourceFile: internal/identity/kube/coretokens_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:29.607Z'
id: TEST-identity-kube
type: test
---
Uses the client-go fake clientset with a reactor that answers TokenRequests on serviceaccounts/token with synthetic tokens and records each request. The first call for ibkr-main asks for ServiceAccount openfief-core in namespace openfief, audience openfief-plugin:ibkr-main and 600 seconds; a call 7m59s later gets the cached token, another instance gets its own, and at two minutes before expiry a new token is requested. An API error or an answer without a token fails and is asked again on the next call, and an empty instance name fails with ErrNoInstance without a request.