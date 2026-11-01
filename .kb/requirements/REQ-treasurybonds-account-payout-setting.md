---
title: 'Each hand-entered bond account must have a payout setting matching the option the bond registry offers: coupons and maturities stay as the account''s cash, or are paid out.'
status: open
text_ref: design-decisions:D-135#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-135-1
  - lane:strict
semantic_text: 'Each hand-entered bond account must have a payout setting matching the option the bond registry offers: coupons and maturities stay as the account''s cash, or are paid out.'
logic_claims:
  - CLAIM-D1992B86C4A73BC8
semantic_clauses:
  - 'Each hand-entered bond account must have a payout setting matching the option the bond registry offers: coupons and maturities stay as the account''s cash, or are paid out.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0b104f345519952c20e93a222fac35cee9b98f2e2713c83ed21fe501a07a3fed
semantic_inventory:
  - claim_key: CLAIM-D1992B86C4A73BC8
    claim_text: 'Each hand-entered bond account must have a payout setting matching the option the bond registry offers: coupons and maturities stay as the account''s cash, or are paid out'
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-135#1
  recorded_at: '2026-11-01T07:39:23.099Z'
id: REQ-treasurybonds-account-payout-setting
type: req
---
Each hand-entered bond account must have a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. The bond registry lets the holder keep coupons and maturities in the registry account or have them paid out, and the setting mirrors that choice. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-135#1
