---
title: The go-webauthn adapter registers and logs in a software passkey and refuses foreign ones
status: active
sourceFile: internal/identity/webauthn/ceremonies_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:38.596Z'
id: TEST-identity-webauthn
type: test
---
Uses a software authenticator that answers creation and request options with attestation none, ES256 and user verification. Checks the registration options (relying party ID is the origin's host, resident key and user verification required, no attestation), a full registration and login returning the credential and counter, and refusal of client data from another origin, an assertion signed for another relying party ID, a user handle that is not the owner's, an unknown credential, and junk JSON. A pass with a broken adapter would need go-webauthn itself to accept a forged assertion.