---
title: Verified With Sigstore Go = true
status: active
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.cosign_bundle
property_key: verified_with_sigstore_go
canonical_key: mcp-kibi-model-requirement:plugins.cosign_bundle:verified_with_sigstore_go:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-51FC27454DF47D8D
claim_text: 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy.'
origin:
  kind: agent
  recorded_at: '2026-10-28T09:17:05.934Z'
id: FACT-PROP-677749B8B15A92CA
type: fact
---
