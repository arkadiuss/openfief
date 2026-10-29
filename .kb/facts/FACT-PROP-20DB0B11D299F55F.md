---
title: Kept Apart From Valuation Prices = true
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.cross_check_prices
property_key: kept_apart_from_valuation_prices
canonical_key: mcp-kibi-model-requirement:valuation.cross_check_prices:kept_apart_from_valuation_prices:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C17FFBB9B4B2F0DD
claim_text: Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:27.230Z'
id: FACT-PROP-20DB0B11D299F55F
type: fact
---
