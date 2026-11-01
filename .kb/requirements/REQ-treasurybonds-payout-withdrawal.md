---
title: When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it.
status: open
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it.
logic_claims:
  - CLAIM-E830DF62E9C06092
semantic_clauses:
  - When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf424a43e3c4dfb4221d06c70ab1019691b3712105aeba8142c609c144a81ffd
semantic_inventory:
  - claim_key: CLAIM-E830DF62E9C06092
    claim_text: When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it
    role: condition
    status: modeled
    span:
      start: 0
      end: 207
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:33.034Z'
id: REQ-treasurybonds-payout-withdrawal
type: req
---
When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it.

## Context

D-135: paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
