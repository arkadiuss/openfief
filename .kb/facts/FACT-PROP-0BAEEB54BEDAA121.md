---
title: Refusal Carries Passkey Check Detail = true
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
subject_key: identity.passkey_check
property_key: refusal_carries_passkey_check_detail
canonical_key: mcp-kibi-model-requirement:identity.passkey_check:refusal_carries_passkey_check_detail:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C484F2B18928522F
claim_text: An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:42.195Z'
id: FACT-PROP-0BAEEB54BEDAA121
type: fact
---
