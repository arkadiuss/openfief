---
title: Coupon Keyed Per Lot And Coupon Date = true
status: active
text_ref: design-decisions:D-143#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-143-4
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: coupon_keyed_per_lot_and_coupon_date
canonical_key: mcp-kibi-model-requirement:plugins.treasury_bonds:coupon_keyed_per_lot_and_coupon_date:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-472431C81B9F1A98
claim_text: The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent.
origin:
  kind: agent
  recorded_at: '2026-11-01T12:25:45.088Z'
id: FACT-PROP-B953AB1484A3A361
type: fact
---
