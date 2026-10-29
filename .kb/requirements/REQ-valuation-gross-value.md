---
title: 'The gross value must be the value before any tax or fee: the market value, or the accrued value of a bond or deposit.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-2
  - lane:strict
text_ref: design-decisions:D-77#2
origin:
  kind: import
  ref: design-decisions:D-77#2
  recorded_at: '2026-10-29T07:58:08.251Z'
semantic_text: 'The gross value must be the value before any tax or fee: the market value, or the accrued value of a bond or deposit.'
logic_claims:
  - CLAIM-78C6802696825E91
semantic_clauses:
  - 'The gross value must be the value before any tax or fee: the market value, or the accrued value of a bond or deposit.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8e82972b0bfcb867d79bc6dcd29ec807f30f79c2e44227cc7d9610bc12341e4b
semantic_inventory:
  - claim_key: CLAIM-78C6802696825E91
    claim_text: 'The gross value must be the value before any tax or fee: the market value, or the accrued value of a bond or deposit'
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-valuation-gross-value
type: req
---
The gross value must be the value before any tax or fee: the market value, or the accrued value of a bond or deposit.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values); this is the first. Charts start on gross, the value broker statements show (REQ-ui-value-switch). The implementation's market valuation (REQ-valuation-market-holding) and its gross_accrued bond basis (REQ-valuation-bond-basis) compute values of this kind today. D-77 replaces D-25's single configurable bond basis, which counted bonds net of tax while stocks stayed gross, so totals mixed two bases (P-40, which D-77 decides). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#2
