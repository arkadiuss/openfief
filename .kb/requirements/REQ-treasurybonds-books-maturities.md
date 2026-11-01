---
title: The treasury-bonds connector must book maturities as it books coupons.
status: open
text_ref: design-decisions:D-135#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-135-3
  - lane:strict
semantic_text: The treasury-bonds connector must book maturities as it books coupons.
logic_claims:
  - CLAIM-AE275C025AEE192D
semantic_clauses:
  - The treasury-bonds connector must book maturities as it books coupons.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 75e0d467b3947ba3b19b349515074dca64131828d252f01e177a22cf422919e1
semantic_inventory:
  - claim_key: CLAIM-AE275C025AEE192D
    claim_text: The treasury-bonds connector must book maturities as it books coupons
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-135#3
  recorded_at: '2026-11-01T07:39:35.754Z'
id: REQ-treasurybonds-books-maturities
type: req
---
The treasury-bonds connector must book maturities as it books coupons.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. D-117 has the treasury-bonds connector deliver coupons computed with the Polish pack's bond rules (REQ-treasurybonds-delivers-coupons), so the core holds no bond logic; maturities follow the same path. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-135#3
