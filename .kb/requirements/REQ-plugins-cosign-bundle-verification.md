---
title: 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy.'
status: open
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
semantic_text: 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy.'
logic_claims:
  - CLAIM-51FC27454DF47D8D
semantic_clauses:
  - 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c39ffa336685f761ca0ba1f851b79312b44a4b05ed29fb10f8896acb0c60532c
semantic_inventory:
  - claim_key: CLAIM-51FC27454DF47D8D
    claim_text: 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy'
    role: normative
    status: modeled
    span:
      start: 0
      end: 442
origin:
  kind: agent
  recorded_at: '2026-10-28T09:17:10.569Z'
id: REQ-plugins-cosign-bundle-verification
type: req
---
The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate's validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image's exact digest by a signer of the trust policy.

## Context

D-17 verifies plugin signatures with cosign; D-56 has the operator do it itself; D-29 forbids home-grown cryptography, so every check is sigstore-go's. OpenFief's CI signs with cosign v3, whose default is a Sigstore bundle stored as an OCI referrer (found through the Referrers API or the referrers tag on GHCR); the older .sig tag format is not read. An attestation by a trusted signer is not a signature. A key signature need not be logged, so a developer's local key works offline (D-57). Refines REQ-AUTO-AC7E986366466A6A and REQ-plugins-operator-verifies-signatures. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> Plugin images must be pinned by digest, and their signatures are verified with cosign.

docs/DESIGN.md D-17

> No home-grown cryptography in M0.

docs/DESIGN.md D-29
