---
title: Cash Converted To Reporting Currency = true
status: active
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.performance_value
property_key: cash_converted_to_reporting_currency
canonical_key: mcp-kibi-model-requirement:views.performance_value:cash_converted_to_reporting_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DC90EE1809242CBB
claim_text: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
origin:
  kind: agent
  recorded_at: '2026-11-02T23:56:07.922Z'
id: FACT-PROP-B8CE7253C96044B6
type: fact
---
