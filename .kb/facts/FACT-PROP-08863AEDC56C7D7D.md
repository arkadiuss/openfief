---
title: Opaque Revocable Hash Only Idle And Absolute Expiry = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-66-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-66#1
fact_kind: property_value
subject_key: identity.session
property_key: opaque_revocable_hash_only_idle_and_absolute_expiry
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:identity.session:opaque_revocable_hash_only_idle_and_absolute_expiry:eq:true
claim_key: CLAIM-56E2628A4496EEF8
claim_text: After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:26:56.111Z'
id: FACT-PROP-08863AEDC56C7D7D
type: fact
---
