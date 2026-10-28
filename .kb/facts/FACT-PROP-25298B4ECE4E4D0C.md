---
title: Deploys Images Without Trusted Signature = false
status: active
text_ref: design-decisions:D-57#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-57-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.operator
property_key: deploys_images_without_trusted_signature
canonical_key: docs-design-md:plugins.operator:deploys_images_without_trusted_signature:eq:false
operator: eq
value_type: bool
value_bool: false
claim_key: CLAIM-7BAFDE8517FB6D01
claim_text: An image without a valid signature from a trusted signer must never be deployed, with no override.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:29.902Z'
id: FACT-PROP-25298B4ECE4E4D0C
type: fact
---
