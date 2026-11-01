---
title: The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow.
status: open
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow.
logic_claims:
  - CLAIM-34351F8E772C2306
semantic_clauses:
  - The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 692345a5c8cfdd9c00bb620db7f91268f1317a950b48852363b5f1dae18465c7
semantic_inventory:
  - claim_key: CLAIM-34351F8E772C2306
    claim_text: The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow
    role: normative
    status: modeled
    span:
      start: 0
      end: 125
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:39.935Z'
id: REQ-views-paid-out-payout-counts-as-return
type: req
---
The time-weighted return must count a paid-out coupon or redemption as return, taking its withdrawal as an external cash flow.

## Context

D-135: the payout still counts as return. D-61's time-weighted return already treats a withdrawal as a cash flow, so no change was needed; the test pins it. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
