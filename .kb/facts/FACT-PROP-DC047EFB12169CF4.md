---
title: Model Portfolio Off Hidden Backtest Exact = true
status: active
text_ref: design-decisions:D-200#19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-19
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: model_portfolio_off_hidden_backtest_exact
canonical_key: mcp-kibi-model-requirement:m0.exit:model_portfolio_off_hidden_backtest_exact:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F903442083B68711
claim_text: While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:53.132Z'
id: FACT-PROP-DC047EFB12169CF4
type: fact
---
