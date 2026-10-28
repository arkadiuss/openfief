---
title: Parses Closes Exactly = true
status: active
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.eodhd_adapter
property_key: parses_closes_exactly
canonical_key: mcp-kibi-model-requirement:valuation.eodhd_adapter:parses_closes_exactly:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DFB9F98A9489C96F
claim_text: The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:29.262Z'
id: FACT-PROP-1994751B710FCD2A
type: fact
---
