---
title: Open Positions Assert Holdings = true
status: active
text_ref: design-decisions:D-155#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-155-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_holding_assertions
property_key: open_positions_assert_holdings
canonical_key: mcp-kibi-model-requirement:ingest.xtb_holding_assertions:open_positions_assert_holdings:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E87C693F776FB8EA
claim_text: 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date.'
origin:
  kind: agent
  recorded_at: '2026-11-01T12:30:04.225Z'
id: FACT-PROP-58EFD03803F04C17
type: fact
---
