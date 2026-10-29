---
title: Only Proves Owner Apart From Session = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-65#2
fact_kind: property_value
subject_key: identity.authenticator
property_key: only_proves_owner_apart_from_session
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:identity.authenticator:only_proves_owner_apart_from_session:eq:true
claim_key: CLAIM-9532B3D91AA7A860
claim_text: Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:26:19.370Z'
id: FACT-PROP-853122DA06C8CC6F
type: fact
---
