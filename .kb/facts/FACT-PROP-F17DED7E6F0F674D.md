---
title: Serves Price Checks = true
status: active
text_ref: docs/DESIGN.md#D-111
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-111
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.read_api
property_key: serves_price_checks
canonical_key: mcp-kibi-model-requirement:views.read_api:serves_price_checks:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-282A98BE4FFD3CB9
claim_text: The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:06.672Z'
id: FACT-PROP-F17DED7E6F0F674D
type: fact
---
