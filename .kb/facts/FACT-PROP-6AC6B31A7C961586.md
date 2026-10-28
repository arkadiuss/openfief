---
title: Asserts Positions And Cash = true
status: active
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ibkr_flex
property_key: asserts_positions_and_cash
canonical_key: mcp-kibi-model-requirement:ingest.ibkr_flex:asserts_positions_and_cash:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4F29CB6236713CAB
claim_text: Each IBKR Flex statement must yield a holding balance assertion per stock or ETF from its open positions, using the SUMMARY row or else the sum of the LOT rows, and a cash balance assertion per currency from its cash report, all dated on the statement's last day.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:59.596Z'
id: FACT-PROP-6AC6B31A7C961586
type: fact
---
