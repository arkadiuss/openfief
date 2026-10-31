---
title: The operator's verifier reports the trust decision as the SignatureVerified condition
status: active
sourceFile: internal/plugins/operator/verify_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:22.459Z'
id: TEST-operator-policy-verifier
type: test
verification_scope: integration
---
Table tests run PolicyVerifier over a fake signed-image source and a policy pinning one synthetic digest: a verified plugin is True/Verified, a CI signature True/SignedByTrustedSigner, an unsigned image False/Unsigned, an unknown signer False/UntrustedSigner, a widened declared manifest False/ManifestMismatch, a missing label False/InvalidManifestLabel and an unreachable registry Unknown/RegistryUnavailable. A cache test checks that a signed image is read once and stays verified while the registry is down, while an unsigned one is read every time. A false pass would be caching a refusal; the unsigned image covers it.