---
title: Raises Item Per Held Assumed Bond Fee = true
status: active
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.checks
property_key: raises_item_per_held_assumed_bond_fee
canonical_key: mcp-kibi-model-requirement:attention.checks:raises_item_per_held_assumed_bond_fee:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0DC16FCEF7E88183
claim_text: Each held treasury bond series whose early-redemption fee is assumed must raise one attention item, its subject the series' ISIN, which must resolve once the owner confirms or overrides the fee or no account holds the series.
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:53.106Z'
id: FACT-PROP-845F1F559F8FAA1A
type: fact
---
