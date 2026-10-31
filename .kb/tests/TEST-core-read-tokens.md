---
title: The read API's verifier accepts the owner's tokens and view plugins' projected tokens, and nothing else
status: active
sourceFile: cmd/core/readtokens_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:13.197Z'
id: TEST-core-read-tokens
type: test
verification_scope: unit
---
Table test of cmd/core's readTokens over a fake owner token check and a fake view plugin caller check: the owner's token and a view plugin's token are accepted; any other token is refused with an error wrapping the read API's ErrInvalidToken; without plugin callers a plugin's token is refused; and when either check fails to run, the error is a failure that does not wrap ErrInvalidToken, so the API answers internal rather than unauthenticated or letting the call through. The real caller check is the AuthService's, tested by TestKubeCallers and TestDevCallers.