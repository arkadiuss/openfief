---
title: On every sync, the treasury-bonds connector must book every coupon due since each lot's purchase.
status: open
text_ref: design-decisions:D-143#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-143-3
  - lane:strict
semantic_text: On every sync, the treasury-bonds connector must book every coupon due since each lot's purchase.
logic_claims:
  - CLAIM-3EB8242A8159C511
semantic_clauses:
  - On every sync, the treasury-bonds connector must book every coupon due since each lot's purchase.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ae644a8bf83e7c8140af1a15273d1a2482f1969f58e809d445f6835b00279962
semantic_inventory:
  - claim_key: CLAIM-3EB8242A8159C511
    claim_text: On every sync, the treasury-bonds connector must book every coupon due since each lot's purchase
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-143#3
  recorded_at: '2026-11-01T12:25:40.500Z'
id: REQ-treasurybonds-coupons-since-purchase-every-sync
type: req
---
On every sync, the treasury-bonds connector must book every coupon due since each lot's purchase.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #35. The decision gives the reason: a lot entered late gets its earlier coupons. D-117 has the connector deliver coupons as interest (REQ-treasurybonds-delivers-coupons), and D-98 books each lot on its real purchase date. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-143#3
