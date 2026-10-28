---
title: Capped At Accrued Interest = true
status: active
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: pl.early_redemption_fee
property_key: capped_at_accrued_interest
canonical_key: mcp-kibi-model-requirement:pl.early_redemption_fee:capped_at_accrued_interest:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F99E6EEB53A988A7
claim_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:59.195Z'
id: FACT-PROP-EC2249412A76A32C
type: fact
---
