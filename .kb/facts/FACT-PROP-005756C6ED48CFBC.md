---
title: Verified Constant Time = true
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.script_token
property_key: verified_constant_time
canonical_key: mcp-kibi-model-requirement:identity.script_token:verified_constant_time:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DC3E19EA220B7EAC
claim_text: Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:24.616Z'
id: FACT-PROP-005756C6ED48CFBC
type: fact
---
