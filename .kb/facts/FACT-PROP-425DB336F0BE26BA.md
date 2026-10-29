---
title: Reports Backtest And Drawdowns When On = true
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
subject_key: portfolio.model_portfolio
property_key: reports_backtest_and_drawdowns_when_on
canonical_key: mcp-kibi-model-requirement:portfolio.model_portfolio:reports_backtest_and_drawdowns_when_on:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-015076512F752AD8
claim_text: While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported.
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:39.074Z'
id: FACT-PROP-425DB336F0BE26BA
type: fact
---
