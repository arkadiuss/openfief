---
title: Books Coupons As Interest = true
status: active
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.coupons
property_key: books_coupons_as_interest
canonical_key: mcp-kibi-model-requirement:treasurybonds.coupons:books_coupons_as_interest:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FADD8F3E5EC210B7
claim_text: In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:35.318Z'
id: FACT-PROP-06C7DA96C5A364FD
type: fact
---
