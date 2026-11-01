---
title: 'The ledger must book a holding redeemed by its issuer as a redemption: the units leave at their redemption value, the tax withheld at source goes to withholding tax and the rest to cash, and a redemption is no sale.'
status: open
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: 'The ledger must book a holding redeemed by its issuer as a redemption: the units leave at their redemption value, the tax withheld at source goes to withholding tax and the rest to cash, and a redemption is no sale.'
logic_claims:
  - CLAIM-BBBCF21212EE8909
semantic_clauses:
  - 'The ledger must book a holding redeemed by its issuer as a redemption: the units leave at their redemption value, the tax withheld at source goes to withholding tax and the rest to cash, and a redemption is no sale.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 65bf1e28bb8ef79944bc013df4acc28e8301aefd929015e11f51531375686521
semantic_inventory:
  - claim_key: CLAIM-BBBCF21212EE8909
    claim_text: 'The ledger must book a holding redeemed by its issuer as a redemption: the units leave at their redemption value, the tax withheld at source goes to withholding tax and the rest to cash, and a redemption is no sale'
    role: normative
    status: modeled
    span:
      start: 0
      end: 214
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:51.598Z'
id: REQ-ledger-redemption-kind
type: req
---
The ledger must book a holding redeemed by its issuer as a redemption: the units leave at their redemption value, the tax withheld at source goes to withholding tax and the rest to cash, and a redemption is no sale.

## Context

D-135 books a maturity as a redemption. A redemption is a kind of its own rather than a sell, so the capital-gains totals (Sales) leave it out: the issuer withholds its tax. Its lots are consumed FIFO like any disposal. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
