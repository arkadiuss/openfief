---
title: Deposit And Buy Per Lot = true
status: active
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.lot_mapping
property_key: deposit_and_buy_per_lot
canonical_key: mcp-kibi-model-requirement:treasurybonds.lot_mapping:deposit_and_buy_per_lot:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-076844F381AA59B7
claim_text: The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:17.883Z'
id: FACT-PROP-CD54057E5E0E98C5
type: fact
---
