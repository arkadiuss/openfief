---
title: Chains Daily Returns Across Flows = true
status: active
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.performance
property_key: chains_daily_returns_across_flows
canonical_key: mcp-kibi-model-requirement:views.performance:chains_daily_returns_across_flows:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-87C6F188421DACA1
claim_text: The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places.
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:49.771Z'
id: FACT-PROP-9567D4A395B2BA4A
type: fact
---
