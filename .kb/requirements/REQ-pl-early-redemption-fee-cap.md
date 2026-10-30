---
title: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
status: closed
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
logic_claims:
  - CLAIM-4B0EDD6B0BB6A79B
semantic_clauses:
  - The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 73e33c857e4e19fcc27744e4900b8505a2158fcdb6ddaf36e0204b4d13fb7d36
semantic_inventory:
  - claim_key: CLAIM-4B0EDD6B0BB6A79B
    claim_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early
    role: exception
    status: modeled
    span:
      start: 0
      end: 306
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:01.244Z'
id: REQ-pl-early-redemption-fee-cap
type: req
---
The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.

## Context

Closed: superseded by REQ-pl-early-redemption-fee-by-wrapper. D-77 replaced the net-if-redeemed value; OKI pays the fee. Modeled by Claude Code for the values work package; the author has not reviewed it yet.
