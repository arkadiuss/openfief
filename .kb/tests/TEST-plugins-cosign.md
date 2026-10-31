---
title: Cosign bundles in a registry are verified offline against the trust root and the trust policy
status: active
sourceFile: internal/plugins/cosign/verifier_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:16:49.591Z'
id: TEST-plugins-cosign
type: test
verification_scope: integration
---
Integration tests push synthetic images and indexes to go-containerregistry's in-memory registry over httptest, with and without the Referrers API, and attach Sigstore v0.3 bundles the way cosign v3 does. Keyless bundles come from sigstore-go's virtual Sigstore (fake Fulcio CA, Rekor and TSA), key bundles from an ephemeral key. Cases: valid CI signature, multi-platform index, catalog-verified publisher, owner key, owner pattern; refused: unsigned, fork identity, other issuer, untrusted key, another digest's signature, a non-signature predicate, an expired certificate, an untrusted Sigstore, an untrusted CA chain, a garbage referrer, a widened declared manifest, no label, differing platform labels, and a registry that tampers with the config blob. The production verifier is shown to require an SCT, and the bundled trust root parses. A false pass would be accepting any bundle; each negative case was checked to fail for its own reason.