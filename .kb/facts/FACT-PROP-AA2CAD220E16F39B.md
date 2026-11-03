---
title: Not Amplified = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.token_review
property_key: not_amplified
canonical_key: mcp-kibi-model-requirement:identity.token_review:not_amplified:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6ACBB95747A6911E
claim_text: The core's browser port must accept only sessions and script tokens, and the core and the egress proxy must send TokenReview only tokens shaped like a JSON Web Token and remember a refusal for a short while, so no caller can make them call the API server at will.
origin:
  kind: agent
  recorded_at: '2026-11-03T00:11:10.974Z'
id: FACT-PROP-AA2CAD220E16F39B
type: fact
---
