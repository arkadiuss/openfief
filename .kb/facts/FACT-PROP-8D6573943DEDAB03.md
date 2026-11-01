---
title: Books From Purchase When Unseeded = true
status: active
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.coupons
property_key: books_from_purchase_when_unseeded
canonical_key: mcp-kibi-model-requirement:treasurybonds.coupons:books_from_purchase_when_unseeded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-39D74F0FC78BBFCB
claim_text: For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:46:15.366Z'
id: FACT-PROP-8D6573943DEDAB03
type: fact
---
