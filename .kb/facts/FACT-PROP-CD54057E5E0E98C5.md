---
title: Deposit And Buy Per Lot = true
status: active
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.lot_mapping
property_key: deposit_and_buy_per_lot
canonical_key: mcp-kibi-model-requirement:treasurybonds.lot_mapping:deposit_and_buy_per_lot:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7A5D33813CE84588
claim_text: The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:17.883Z'
id: FACT-PROP-CD54057E5E0E98C5
type: fact
---
