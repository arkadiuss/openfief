---
title: The SDK's JWKS verifier accepts only the core's unexpired token for this instance
status: active
sourceFile: sdk/go/openfief/verifier_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:20.101Z'
id: TEST-sdk-go-jwks-verifier
type: test
---
Table tests in package openfief_test generate RSA and ECDSA keys locally, write a JWKS and an issuer file into a temporary directory as the operator's ConfigMap mount holds them, and sign synthetic ServiceAccount tokens with go-jose. They check that RS256 and ES256 tokens of the core's subject, issuer and this instance's audience pass, and that another audience, the core's own audience, another subject, another issuer, an expired token, a token not valid yet, a token without expiry, an unknown key, another key under a known key ID, no key ID, alg none, HMAC keyed with the public key and garbage are refused with ErrNotCore and never echo the token. Key-set cases cover missing files, an empty issuer, bad JSON, empty sets, symmetric, encryption-only and ID-less keys and an unknown key type beside a good key. A rotation test rewrites the files and checks the new key is used at once. RequireCore with the verifier answers 401 to other tokens. A false pass would be a verifier that skips the signature; the wrong-key cases guard it.