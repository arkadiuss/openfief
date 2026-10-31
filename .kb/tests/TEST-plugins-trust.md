---
title: Signer identities match exactly and the trust policy decides verified, unverified or refused
status: active
sourceFile: internal/plugins/trust_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:18.475Z'
id: TEST-plugins-trust
type: test
verification_scope: unit
---
Table tests in package plugins_test build synthetic keyless identities, GitHub Actions identities and fresh ECDSA public keys. They check that identity validation rejects mixed, empty, non-https and malformed identities; that matching needs the same issuer and the exact subject or a pattern matching the whole subject, or the same key regardless of PEM whitespace; that OpenFief's CI identity rejects a fork and another branch; and that TrustPolicy.Decide returns Verified only for a catalog-pinned digest signed by its publisher, Unverified for any other trusted signer, and ErrUnsigned, ErrUntrustedSigner or ErrManifestMismatch otherwise. A false pass would be an unanchored pattern; two cases append and prepend text to the subject.