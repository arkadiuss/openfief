---
title: The trust policy refuses an image signed only by untrusted signers apart from an unsigned one, naming them
status: active
verification_scope: unit
sourceFile: internal/plugins/untrusted_test.go
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:25.235Z'
id: TEST-plugins-untrusted-signer
type: test
---
Table tests in package plugins_test run TrustPolicy.Decide over synthetic signed images: no signature at all gives ErrUnsigned; only a keyless signer nobody trusts, only an untrusted key signed twice, or both give an UntrustedSignerError matching ErrUntrustedSigner that names the image and each signer once, keyless as subject and issuer and a key as its SHA-256 fingerprint; and an untrusted signature next to the owner's key still decides Unverified by the owner's signer. A second test checks the error's message, which the operator writes into the Plugin's status. A false pass would be an untrusted signature that admits an image; the last case asserts the trusted signer decides.