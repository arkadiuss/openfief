---
title: Change Revalues All = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.bond_basis_setting
property_key: change_revalues_all
canonical_key: mcp-kibi-model-requirement:valuation.bond_basis_setting:change_revalues_all:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7CC9DCDD33B4D25A
claim_text: Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:53.703Z'
id: FACT-PROP-F3E3379FA13E702D
type: fact
---
