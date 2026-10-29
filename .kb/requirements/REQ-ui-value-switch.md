---
title: Each chart must have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
status: closed
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-6
  - lane:strict
text_ref: design-decisions:D-77#6
origin:
  kind: import
  ref: design-decisions:D-77#6
  recorded_at: '2026-10-29T08:01:29.247Z'
semantic_text: Each chart must have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
logic_claims:
  - CLAIM-2BE87C34BAE1F3CD
semantic_clauses:
  - Each chart must have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: bc03d396c4b9043efcfdc30a647297b59f0c6263a3d4f91ab4b3b6059f701f5a
semantic_inventory:
  - claim_key: CLAIM-2BE87C34BAE1F3CD
    claim_text: Each chart must have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines
    role: normative
    status: modeled
    span:
      start: 0
      end: 164
id: REQ-ui-value-switch
type: req
---
Each chart must have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.

## Why this changed

The author's decision D-77 replaces D-25, under which bonds counted in totals at their net-if-redeemed-today value by default. Under D-77 the value shown first is gross, the value broker statements show, and the owner switches to net or liquidation per chart. REQ-AUTO-E10BE8CE13E67657 stated D-25's default, and D-25 has left docs/DESIGN.md, so this requirement replaces that default; REQ-valuation-liquidation-value replaces the value it defined.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values), and tables show all three side by side (REQ-ui-three-values-tables). The value-over-time chart is D-30's, and the built-in UI's overview draws it today as the portfolio's time-weighted return index (REQ-ui-overview). D-25 counted bonds net of tax while stocks stayed gross, so totals mixed two bases (P-40, which D-77 decides). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#6
