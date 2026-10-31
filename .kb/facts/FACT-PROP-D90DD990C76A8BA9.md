---
title: Leg In One Confirmed Transfer = true
status: active
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.transfer
property_key: leg_in_one_confirmed_transfer
canonical_key: mcp-kibi-model-requirement:ledger.transfer:leg_in_one_confirmed_transfer:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-15A99C908FFCB459
claim_text: Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race.
origin:
  kind: agent
  recorded_at: '2026-10-30T06:06:08.429Z'
id: FACT-PROP-D90DD990C76A8BA9
type: fact
---
