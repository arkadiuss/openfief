---
title: Values Holdings And Cash = true
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
subject_key: views.allocation
property_key: values_holdings_and_cash
canonical_key: mcp-kibi-model-requirement:views.allocation:values_holdings_and_cash:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2F1750F1D6817487
claim_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:26:06.997Z'
id: FACT-PROP-AC3AA2066CD732AD
type: fact
---
