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
subject_key: views.allocation_inputs
property_key: cash_converted_to_reporting_currency
canonical_key: mcp-kibi-model-requirement:views.allocation_inputs:cash_converted_to_reporting_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-41E7208C7B0F0EF3
claim_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate.
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:56.625Z'
id: FACT-PROP-73F4472B2750EC3A
type: fact
---
