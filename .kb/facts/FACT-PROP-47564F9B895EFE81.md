---
title: Rebalance Interval = yearly
status: active
text_ref: design-decisions:D-53#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-53-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: portfolio.backtest
property_key: rebalance_interval
canonical_key: docs-design-md:portfolio.backtest:rebalance_interval:eq:yearly
operator: eq
value_type: string
value_string: yearly
claim_key: CLAIM-3D89D9DD3B0B389D
claim_text: The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:05.624Z'
id: FACT-PROP-47564F9B895EFE81
type: fact
---
