---
title: Other Source Price Checks = true
status: active
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.market_data
property_key: other_source_price_checks
canonical_key: mcp-kibi-model-requirement:valuation.market_data:other_source_price_checks:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C9CD77A25D0B21EB
claim_text: Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding.
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:57.028Z'
id: FACT-PROP-18337A3CC07A5B3D
type: fact
---
