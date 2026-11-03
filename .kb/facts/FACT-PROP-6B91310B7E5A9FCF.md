---
title: Need Recent Passkey Check = true
status: active
text_ref: design-decisions:D-183#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-183-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.protected_actions
property_key: need_recent_passkey_check
canonical_key: mcp-kibi-model-requirement:identity.protected_actions:need_recent_passkey_check:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E81DC28A7B86734E
claim_text: Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session.
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:23.058Z'
id: FACT-PROP-6B91310B7E5A9FCF
type: fact
---
