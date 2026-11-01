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
Integration tests push synthetic images and indexes to go-containerregistry's in-memory registry over httptest, with and without the Referrers API, and attach Sigstore v0.3 bundles the way cosign v3 does. Keyless bundles come from sigstore-go's virtual Sigstore (fake Fulcio CA, Rekor and TSA), key bundles from an ephemeral key. Cases: valid CI signature, multi-platform index, catalog-verified publisher, owner key, owner pattern; refused as UntrustedSigner and named (D-169): a fork's identity and another issuer by subject and issuer, an untrusted key by its fingerprint; a trusted signature next to an untrusted one still decides; refused as unsigned and naming nobody: no signature, another digest's signature, a non-signature predicate by a trusted signer or by an unknown key, a key statement over another image, a forged hint of a trusted key, an expired certificate, an untrusted Sigstore, an untrusted CA chain, a garbage referrer; and a widened declared manifest, no label, differing platform labels, and a registry that tampers with the config blob. Every named signer must appear in the refusal. The production verifier is shown to require an SCT, and the bundled trust root parses. A false pass would be accepting any bundle, or naming a signature that isn't valid; each negative case was checked to fail for its own reason.