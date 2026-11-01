---
title: The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once.
status: open
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once.
logic_claims:
  - CLAIM-2825F44C127438A9
semantic_clauses:
  - The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c0d6c7914853d4e259be4dcf4f80e20865b9a12434d12f15ab8d256db3e7cfc4
semantic_inventory:
  - claim_key: CLAIM-2825F44C127438A9
    claim_text: The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once
    role: normative
    status: modeled
    span:
      start: 0
      end: 285
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:24.404Z'
id: REQ-treasurybonds-maturity-redemption
type: req
---
The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once.

## Context

D-135: the connector books maturities as it books coupons. The last period's interest of a coupon series comes with the redemption; D-100 taxes the exchange discount with the interest. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
