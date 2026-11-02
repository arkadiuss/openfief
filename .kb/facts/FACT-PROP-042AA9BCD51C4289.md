---
title: Fees Split With Units = true
status: active
text_ref: docs/DESIGN.md#D-150
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-150
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.lot_buy_fees
property_key: fees_split_with_units
canonical_key: mcp-kibi-model-requirement:ledger.lot_buy_fees:fees_split_with_units:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A5E57AC2F5F22E73
claim_text: A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest.
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:34.476Z'
id: FACT-PROP-042AA9BCD51C4289
type: fact
---
