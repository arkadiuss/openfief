---
title: Sums Lot Values On Basis = true
status: active
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.bond_holding
property_key: sums_lot_values_on_basis
canonical_key: mcp-kibi-model-requirement:valuation.bond_holding:sums_lot_values_on_basis:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-97670979153E46CE
claim_text: A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:59.095Z'
id: FACT-PROP-D10588F591F6C05A
type: fact
---
