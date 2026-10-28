---
title: Empty Range Is Not An Error = true
status: active
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.fx_rate_provider
property_key: empty_range_is_not_an_error
canonical_key: mcp-kibi-model-requirement:valuation.fx_rate_provider:empty_range_is_not_an_error:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E7D2F3F724AC05C7
claim_text: A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:27.511Z'
id: FACT-PROP-33DFA6F1083EAE4F
type: fact
---
