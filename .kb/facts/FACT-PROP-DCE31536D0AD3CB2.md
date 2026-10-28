---
title: Revocable = true
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
property_key: revocable
canonical_key: mcp-kibi-model-requirement:identity.script_token:revocable:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4BC942053859F808
claim_text: Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:29.202Z'
id: FACT-PROP-DCE31536D0AD3CB2
type: fact
---
