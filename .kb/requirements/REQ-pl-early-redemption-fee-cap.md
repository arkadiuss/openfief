---
title: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity.
logic_claims:
  - CLAIM-F99E6EEB53A988A7
semantic_clauses:
  - The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 27850847cc0335f89415aae1a6349111b9ca80348140a97b10646d2dbecca37e
semantic_inventory:
  - claim_key: CLAIM-F99E6EEB53A988A7
    claim_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity
    role: exception
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:01.244Z'
id: REQ-pl-early-redemption-fee-cap
type: req
---
The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, and no fee at or after maturity.

## Context

D-25 makes the default bond value net of the early-redemption fee. The fee per bond is series data (it rose for issues from September 2024). The cap follows the issue letters as summarized by the bond offer and secondary sources; for ROR, DOR and COI the cap applies only in the first interest period because earlier coupons were already paid out. Refines REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet. obligacjeskarbowe.pl and gov.pl were not reachable from the agent's sandbox, so the rule was checked only against search-result excerpts of issue letters and secondary calculators.

## Source

> the default is net-if-redeemed-today (after the early-redemption fee and Belka tax, no tax inside IKE)

docs/DESIGN.md D-25
