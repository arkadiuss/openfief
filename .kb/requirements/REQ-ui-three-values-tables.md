---
title: Tables must show the gross, net and liquidation values side by side.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-5
  - lane:strict
text_ref: design-decisions:D-77#5
origin:
  kind: import
  ref: design-decisions:D-77#5
  recorded_at: '2026-10-29T08:00:44.609Z'
semantic_text: Tables must show the gross, net and liquidation values side by side.
logic_claims:
  - CLAIM-CACE31861CE1D19F
semantic_clauses:
  - Tables must show the gross, net and liquidation values side by side.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4324498a078876e1062f0be12844d4dbb932b776c6edea2bc15643a29799c125
semantic_inventory:
  - claim_key: CLAIM-CACE31861CE1D19F
    claim_text: Tables must show the gross, net and liquidation values side by side
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
id: REQ-ui-three-values-tables
type: req
---
Tables must show the gross, net and liquidation values side by side.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values), defined in REQ-valuation-gross-value, REQ-valuation-net-value and REQ-valuation-liquidation-value; tables show all three, while charts switch between them (REQ-ui-value-switch). The built-in UI shows one value per holding and total today (REQ-ui-overview, REQ-ui-account-pages). D-77 replaces D-25's single configurable bond basis, which counted bonds net of tax while stocks stayed gross, so totals mixed two bases (P-40, which D-77 decides). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#5
