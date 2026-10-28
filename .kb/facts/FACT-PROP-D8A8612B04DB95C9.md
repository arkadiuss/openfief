---
title: Values Targets In Reporting Currency = true
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
subject_key: portfolio.backtest
property_key: values_targets_in_reporting_currency
canonical_key: mcp-kibi-model-requirement:portfolio.backtest:values_targets_in_reporting_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A33ED550CC477AF2
claim_text: Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:38.353Z'
id: FACT-PROP-D8A8612B04DB95C9
type: fact
---
