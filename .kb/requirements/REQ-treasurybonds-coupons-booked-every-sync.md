---
title: On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync.
status: open
text_ref: docs/DESIGN.md#D-143
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-143
  - lane:strict
semantic_text: On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync.
logic_claims:
  - CLAIM-E9CA7909BCE940CF
semantic_clauses:
  - On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b7bc1256623fb084fd57cfce8e623ec4306e23d7f0180e45b5191779ac375227
semantic_inventory:
  - claim_key: CLAIM-E9CA7909BCE940CF
    claim_text: On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync
    role: normative
    status: modeled
    span:
      start: 0
      end: 328
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:29.454Z'
id: REQ-treasurybonds-coupons-booked-every-sync
type: req
---
On every sync, the treasury-bonds connector must book as interest each coupon that fell due on or before today since the purchase of an open lot the request names, keyed per lot and coupon date, with Belka tax withheld by the account's tax wrapper, and leave a coupon whose index reading is missing as a problem for a later sync.

## Context

D-143 has the treasury-bonds connector book every coupon due since each lot's purchase on every sync, not only those due since its last sync (D-117), so a lot entered late gets its earlier coupons; keys per lot and date keep it idempotent. Interest (D-101) and Belka tax per wrapper (D-76) are the existing rules. Which coupons an account seeded with opening lots gets is REQ-treasurybonds-payouts-from-purchase.

## Why this changed

The author decided D-143, so it replaces REQ-treasurybonds-coupons-as-interest: that requirement described only a transactions sync and only coupons after the account's go-live, while D-143 books every coupon since each lot's purchase on every sync. It is closed.

## Source

> **[D-143]** Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

docs/DESIGN.md#D-143
