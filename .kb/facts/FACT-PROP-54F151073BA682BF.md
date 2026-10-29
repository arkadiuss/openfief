---
title: Rebalances At Year End = true
status: active
text_ref: docs/DESIGN.md#D-53
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-53
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: portfolio.backtest
property_key: rebalances_at_year_end
canonical_key: mcp-kibi-model-requirement:portfolio.backtest:rebalances_at_year_end:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E707A47D88E7638E
claim_text: The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:42.792Z'
id: FACT-PROP-54F151073BA682BF
type: fact
---
