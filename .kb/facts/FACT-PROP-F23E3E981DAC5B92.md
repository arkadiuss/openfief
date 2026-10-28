---
title: Sums Per Date In Pln And Per Currency = true
status: active
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.valuation_totals
property_key: sums_per_date_in_pln_and_per_currency
canonical_key: mcp-kibi-model-requirement:views.valuation_totals:sums_per_date_in_pln_and_per_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1BE71A61676C6B85
claim_text: Valuation totals must be one per date, summing the stored valuations dated on it exactly in PLN and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:25.393Z'
id: FACT-PROP-F23E3E981DAC5B92
type: fact
---
