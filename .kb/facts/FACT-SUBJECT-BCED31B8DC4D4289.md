---
title: Plugins Cosign Bundle
status: active
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.cosign_bundle
canonical_key: plugins.cosign_bundle
claim_key: CLAIM-51FC27454DF47D8D
claim_text: 'The operator must count an image signature only when sigstore-go verifies a cosign bundle attached to the image as an OCI referrer: for a keyless signature a certificate chain to a trusted Fulcio CA, an SCT, a trusted transparency log entry and an observer timestamp within the certificate''s validity; for a key signature a trusted key; and in both cases the cosign sign predicate over the image''s exact digest by a signer of the trust policy.'
origin:
  kind: agent
  recorded_at: '2026-10-28T09:17:02.146Z'
id: FACT-SUBJECT-BCED31B8DC4D4289
type: fact
---
