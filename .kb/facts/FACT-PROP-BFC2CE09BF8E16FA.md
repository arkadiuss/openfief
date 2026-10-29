---
title: Flags Difference Beyond Threshold = true
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
subject_key: valuation.price_check
property_key: flags_difference_beyond_threshold
canonical_key: mcp-kibi-model-requirement:valuation.price_check:flags_difference_beyond_threshold:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-03E78053777F2006
claim_text: A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:19.192Z'
id: FACT-PROP-BFC2CE09BF8E16FA
type: fact
---
