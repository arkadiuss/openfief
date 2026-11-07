---
title: Reads Ministry Spreadsheet = true
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
subject_key: valuation.mfbonds
property_key: reads_ministry_spreadsheet
canonical_key: mcp-kibi-model-requirement:valuation.mfbonds:reads_ministry_spreadsheet:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C96F2C9F19C1176B
claim_text: The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl.
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:37.492Z'
id: FACT-PROP-0E1B5ABEB8DAE1FC
type: fact
---
