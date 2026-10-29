---
title: Is Opaque Hashed And Expiring = true
status: active
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.session
property_key: is_opaque_hashed_and_expiring
canonical_key: mcp-kibi-model-requirement:identity.session:is_opaque_hashed_and_expiring:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-680DD6D38E718879
claim_text: After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:07.899Z'
id: FACT-PROP-292FC52C2DCE4290
type: fact
---
