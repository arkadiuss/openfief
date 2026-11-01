---
title: The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent.
status: open
text_ref: design-decisions:D-143#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-143-4
  - lane:strict
semantic_text: The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent.
logic_claims:
  - CLAIM-472431C81B9F1A98
semantic_clauses:
  - The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5350cdf1355463e94541adaf4147ff6e476ada0eb71f5dbaa8aea0bcd90406d1
semantic_inventory:
  - claim_key: CLAIM-472431C81B9F1A98
    claim_text: The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-143#4
  recorded_at: '2026-11-01T12:25:48.102Z'
id: REQ-treasurybonds-coupons-keyed-per-lot-and-date
type: req
---
The treasury-bonds connector must key each coupon it books per lot and coupon date, so a sync stays idempotent.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #35. D-18 makes re-imports idempotent, keyed by source plus the institution's own ID; booking every coupon since purchase on each sync relies on that key. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-143#4
