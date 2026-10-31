---
title: Records Fx Rates From Earliest Foreign Lot = true
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
subject_key: demo.loader
property_key: records_fx_rates_from_earliest_foreign_lot
canonical_key: mcp-kibi-model-requirement:demo.loader:records_fx_rates_from_earliest_foreign_lot:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6918EAB86017CFB8
claim_text: The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:17:17.301Z'
id: FACT-PROP-F955AEBA4D601505
type: fact
---
