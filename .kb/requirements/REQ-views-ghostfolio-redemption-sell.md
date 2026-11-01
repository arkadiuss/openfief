---
title: The Ghostfolio view must project a redemption as a SELL of its units at the redemption value, with the tax withheld as its fee.
status: open
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: The Ghostfolio view must project a redemption as a SELL of its units at the redemption value, with the tax withheld as its fee.
logic_claims:
  - CLAIM-3E0CD003DA7A9161
semantic_clauses:
  - The Ghostfolio view must project a redemption as a SELL of its units at the redemption value, with the tax withheld as its fee.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 23cc3970ae46b78b42f9f9e286b0b62c3a35fdd5d87255aa7eaf87587c6d2e8c
semantic_inventory:
  - claim_key: CLAIM-3E0CD003DA7A9161
    claim_text: The Ghostfolio view must project a redemption as a SELL of its units at the redemption value, with the tax withheld as its fee
    role: normative
    status: modeled
    span:
      start: 0
      end: 126
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:58.100Z'
id: REQ-views-ghostfolio-redemption-sell
type: req
---
The Ghostfolio view must project a redemption as a SELL of its units at the redemption value, with the tax withheld as its fee.

## Context

Ghostfolio has no redemption activity; a SELL keeps its holdings and cash right. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
