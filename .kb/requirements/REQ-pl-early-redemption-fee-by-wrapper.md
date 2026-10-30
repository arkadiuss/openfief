---
title: A bond lot's early-redemption fee must be its series' fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, none at or after maturity and none for OTS, waived in IKE and IKZE and charged in a taxable account and in OKI.
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: A bond lot's early-redemption fee must be its series' fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, none at or after maturity and none for OTS, waived in IKE and IKZE and charged in a taxable account and in OKI.
logic_claims:
  - CLAIM-B8AB1C971D4659A6
semantic_clauses:
  - A bond lot's early-redemption fee must be its series' fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, none at or after maturity and none for OTS, waived in IKE and IKZE and charged in a taxable account and in OKI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 99effc5fe9b60c29fe4581bb087f05c9b6e3f64d0bc8dff00a2b6c1b14351c33
semantic_inventory:
  - claim_key: CLAIM-B8AB1C971D4659A6
    claim_text: A bond lot's early-redemption fee must be its series' fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, none at or after maturity and none for OTS, waived in IKE and IKZE and charged in a taxable account and in OKI
    role: exception
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:22.003Z'
id: REQ-pl-early-redemption-fee-by-wrapper
type: req
---
A bond lot's early-redemption fee must be its series' fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, none at or after maturity and none for OTS, waived in IKE and IKZE and charged in a taxable account and in OKI.

## Context

Supersedes REQ-pl-early-redemption-fee-cap, which spoke of the net-if-redeemed value D-77 replaced. The issue letters waive the fee on withdrawals from IKE and IKZE only, so OKI pays it: this package's narrow reading of D-76, which doesn't say. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-pl-early-redemption-fee-cap: D-77 replaced the net-if-redeemed value; OKI pays the fee. The superseded requirement is closed.

## Source

> **[D-77]** Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper: Gross: before any tax or fee; market value, or the accrued value of a bond or deposit. Net: what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees. Liquidation: the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

docs/DESIGN.md#D-77
