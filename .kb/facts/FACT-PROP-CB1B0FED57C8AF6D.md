---
title: Booked Since Purchase Every Sync = true
status: active
text_ref: docs/DESIGN.md#D-143
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-143
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.coupons
property_key: booked_since_purchase_every_sync
canonical_key: mcp-kibi-model-requirement:treasurybonds.coupons:booked_since_purchase_every_sync:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E9CA7909BCE940CF
claim_text: On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync.
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:26.258Z'
id: FACT-PROP-CB1B0FED57C8AF6D
type: fact
---
