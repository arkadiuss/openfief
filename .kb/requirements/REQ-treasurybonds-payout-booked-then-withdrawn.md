---
title: When a bond account's coupons and maturities are paid out, each must be booked as interest or a redemption and then a withdrawal of the same amount.
status: open
text_ref: design-decisions:D-135#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-135-2
  - lane:strict
semantic_text: When a bond account's coupons and maturities are paid out, each must be booked as interest or a redemption and then a withdrawal of the same amount.
logic_claims:
  - CLAIM-8122CD40D61B24E5
semantic_clauses:
  - When a bond account's coupons and maturities are paid out, each must be booked as interest or a redemption and then a withdrawal of the same amount.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0fec40ed2ac2f24a54b520352ae8ad120614410ab0dbffd84ba0e73204f5810d
semantic_inventory:
  - claim_key: CLAIM-8122CD40D61B24E5
    claim_text: When a bond account's coupons and maturities are paid out, each must be booked as interest or a redemption and then a withdrawal of the same amount
    role: condition
    status: modeled
    span:
      start: 0
      end: 147
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-135#2
  recorded_at: '2026-11-01T07:39:29.264Z'
id: REQ-treasurybonds-payout-booked-then-withdrawn
type: req
---
When a bond account's coupons and maturities are paid out, each must be booked as interest or a redemption and then a withdrawal of the same amount.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. The decision gives the reason: the account's cash then matches the registry, and the payout still counts as return. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-135#2
