---
title: Asserts Open Positions Holdings = true
status: active
text_ref: docs/DESIGN.md#D-155
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-155
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_mapper
property_key: asserts_open_positions_holdings
canonical_key: mcp-kibi-model-requirement:ingest.xtb_mapper:asserts_open_positions_holdings:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6FBC3F20F446E60C
claim_text: The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it.
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:40.132Z'
id: FACT-PROP-1C957452EA3CE64F
type: fact
---
