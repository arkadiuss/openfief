---
title: In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.
status: closed
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
semantic_text: In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.
logic_claims:
  - CLAIM-FADD8F3E5EC210B7
semantic_clauses:
  - In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0c4cc7b324df028dc5793055133376990f5e539b8287c4264def326d94cf1ddd
semantic_inventory:
  - claim_key: CLAIM-FADD8F3E5EC210B7
    claim_text: In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:37.572Z'
id: REQ-treasurybonds-coupons-as-interest
type: req
---
In a transactions sync, the treasury-bonds connector must book each coupon that fell due after the account's go-live and on or before today on an open lot the request names as interest, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper.

## Context

D-117 sends the holdings and open lots with every sync and D-101 books coupons as interest. Coupons on or before go-live are history the opening lots hold (D-98); a deleted lot is no longer open (D-100). The connector delivers every coupon due on every sync, which covers those since its last sync and stays idempotent by key. A coupon whose index reading is not published waits as a problem. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Each sync request to a connector carries the holdings and open lots of the accounts bound to that instance, as D-93's request carries the instruments to price. The treasury-bonds connector uses them to deliver each coupon that fell due since its last sync as interest (D-101), computed with the Polish pack's bond rules, so the core holds no bond logic.

docs/DESIGN.md#D-117
