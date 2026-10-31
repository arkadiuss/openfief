---
title: Only an issuer and public RSA or EC keys with IDs are accepted for publishing
status: active
sourceFile: internal/plugins/signing_keys_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:22.040Z'
id: TEST-plugins-signing-keys
type: test
verification_scope: unit
---
Table tests in package plugins_test feed ParseSigningKeys synthetic OpenID configurations and key sets in the shape the API server serves them. An RSA plus EC set is accepted and compacted, keeping the issuer; a missing issuer, an issuer with a newline, a configuration or set that is not JSON, an empty set, a symmetric key, a private RSA key and a key without an ID are refused with ErrInvalidSigningKeys. Data is checked to hold the set under jwks.json and the issuer under issuer. A false pass would be publishing private key material; the private key case guards it.