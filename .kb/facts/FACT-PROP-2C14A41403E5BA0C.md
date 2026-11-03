---
title: Revokes Sessions It Signed In = true
status: active
text_ref: docs/DESIGN.md#D-183
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-183
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.passkey_removal
property_key: revokes_sessions_it_signed_in
canonical_key: mcp-kibi-model-requirement:identity.passkey_removal:revokes_sessions_it_signed_in:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6F599805071778E7
claim_text: Removing a passkey must revoke every session that passkey signed in, in the same transaction.
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:51.806Z'
id: FACT-PROP-2C14A41403E5BA0C
type: fact
---
