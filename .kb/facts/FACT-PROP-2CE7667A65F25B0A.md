---
title: Cost Split Conserves Money = true
status: active
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.lot
property_key: cost_split_conserves_money
canonical_key: mcp-kibi-model-requirement:ledger.lot:cost_split_conserves_money:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1F085A54BEC551B4
claim_text: Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:23.072Z'
id: FACT-PROP-2CE7667A65F25B0A
type: fact
---
