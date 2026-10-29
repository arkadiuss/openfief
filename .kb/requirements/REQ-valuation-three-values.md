---
title: Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-1
  - lane:strict
text_ref: design-decisions:D-77#1
origin:
  kind: import
  ref: design-decisions:D-77#1
  recorded_at: '2026-10-29T07:57:10.742Z'
semantic_text: Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper.
logic_claims:
  - CLAIM-94008AF0009FB3B6
semantic_clauses:
  - Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6fa83977e58274a9bf52df2ad29cef6bf3df15b9053cfd4263defed09b1d6b7d
semantic_inventory:
  - claim_key: CLAIM-94008AF0009FB3B6
    claim_text: Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper
    role: normative
    status: modeled
    span:
      start: 0
      end: 159
id: REQ-valuation-three-values
type: req
---
Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper.

## Why this changed

The author's decision D-77 replaces D-25, which made the bond value that counts in totals one configurable basis, net-if-redeemed-today by default. Under D-77 there is no single configurable basis: every holding, account and total has all three values, gross, net and liquidation, and D-25 has left docs/DESIGN.md. REQ-AUTO-0C90AD924924F4F6 required that basis to be configurable, so this requirement replaces it.

## Context

D-25 counted bonds net of tax while stocks stayed gross, so totals mixed two bases; P-40 proposed an after-tax liquidation value, and D-77 decides it. D-70 makes the jurisdiction pack a set of pure functions inside the core, and D-76 gives every account a tax wrapper (REQ-ledger-account-tax-wrapper). REQ-valuation-gross-value, REQ-valuation-net-value and REQ-valuation-liquidation-value define the three values; D-36 now stores all three in each valuation (REQ-decision-d36-9-three-values-stored). The implementation still values bonds on one configured basis (REQ-valuation-bond-basis, REQ-valuation-bond-holding). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#1
