---
title: Serves Price Checks = true
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
subject_key: views.read_api
property_key: serves_price_checks
canonical_key: mcp-kibi-model-requirement:views.read_api:serves_price_checks:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6E8AD9D00747FAFD
claim_text: The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:06.672Z'
id: FACT-PROP-F17DED7E6F0F674D
type: fact
---
