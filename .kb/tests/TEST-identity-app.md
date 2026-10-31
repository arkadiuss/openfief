---
title: Issuing, verifying, revoking and listing script tokens through the use cases
status: active
sourceFile: internal/identity/app/tokens_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:03.788Z'
id: TEST-identity-app
type: test
verification_scope: unit
---
The use cases run over an in-memory store with a settable clock and a fixed random source. Issue returns the read-scoped token with the expected ID, expiry and secret, and the store holds only the SHA-256 hash of the key; failures of validation, the random source or the store return no secret and store nothing. Verify accepts the issued secret just before expiry and refuses it at expiry, after revocation, for another scope, malformed, with an unknown ID or another key, each as ErrInvalidToken, while a store failure is never reported as an invalid token and no error contains the secret. Revoke keeps the first revocation time and refuses unknown and malformed IDs.