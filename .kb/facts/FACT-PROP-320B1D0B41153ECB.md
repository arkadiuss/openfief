---
title: Seeds From Open Positions = true
status: active
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_connector
property_key: seeds_from_open_positions
canonical_key: mcp-kibi-model-requirement:ingest.xtb_connector:seeds_from_open_positions:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-80FF230DDDA849F0
claim_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account''s currency, on the report''s date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN.'
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:19.371Z'
id: FACT-PROP-320B1D0B41153ECB
type: fact
---
