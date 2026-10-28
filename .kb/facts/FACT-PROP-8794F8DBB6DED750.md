---
title: Deploys Only Verified Images = true
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
subject_key: operator.image_signature
property_key: deploys_only_verified_images
canonical_key: mcp-kibi-model-requirement:operator.image_signature:deploys_only_verified_images:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5BC1096F3F857964
claim_text: The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:21.297Z'
id: FACT-PROP-8794F8DBB6DED750
type: fact
---
