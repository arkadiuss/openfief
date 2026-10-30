---
title: Sums Lot Values = true
status: active
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.bond_holding
property_key: sums_lot_values
canonical_key: mcp-kibi-model-requirement:valuation.bond_holding:sums_lot_values:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-63D0850AA2500FBA
claim_text: A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded.
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:58.458Z'
id: FACT-PROP-D0A08ACC4AD65944
type: fact
---
