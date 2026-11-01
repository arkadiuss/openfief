---
title: For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch.
logic_claims:
  - CLAIM-39D74F0FC78BBFCB
semantic_clauses:
  - For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e859b8f28d8703dfb71c28671ff16e3a1229699c18d47bce1ab14847507f9c01
semantic_inventory:
  - claim_key: CLAIM-39D74F0FC78BBFCB
    claim_text: For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch
    role: normative
    status: modeled
    span:
      start: 0
      end: 231
origin:
  kind: agent
  recorded_at: '2026-11-01T07:46:18.385Z'
id: REQ-treasurybonds-payouts-from-purchase
type: req
---
For an account never seeded, the treasury-bonds connector must book every coupon and maturity since purchase in an opening sync batch, and for an account seeded with opening lots only those after its go-live in a transactions batch.

## Context

D-98 rewritten: the account's history starts at its first purchase, so nothing before is held by opening lots. Such accounts are never seeded, so the scheduler syncs them in opening mode (D-95); booking each account in one mode keeps an opening batch from seeding a seeded account twice. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

docs/DESIGN.md#D-98
