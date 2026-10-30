---
title: The built-in UI must show gross, net and liquidation values side by side for the portfolio's total, each account and each holding, a missing value saying so with the read API's reason, and must mark a holding or lot no longer updated with the removed instances behind it.
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: The built-in UI must show gross, net and liquidation values side by side for the portfolio's total, each account and each holding, a missing value saying so with the read API's reason, and must mark a holding or lot no longer updated with the removed instances behind it.
logic_claims:
  - CLAIM-CAFC030443491A7A
semantic_clauses:
  - The built-in UI must show gross, net and liquidation values side by side for the portfolio's total, each account and each holding, a missing value saying so with the read API's reason, and must mark a holding or lot no longer updated with the removed instances behind it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 67c9f1818370287942f164ebfd7f4e56c3fb7510d5b8307669bd0fd78dd109f7
semantic_inventory:
  - claim_key: CLAIM-CAFC030443491A7A
    claim_text: The built-in UI must show gross, net and liquidation values side by side for the portfolio's total, each account and each holding, a missing value saying so with the read API's reason, and must mark a holding or lot no longer updated with the removed instances behind it
    role: normative
    status: modeled
    span:
      start: 0
      end: 270
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:28.354Z'
id: REQ-ui-three-values-shown
type: req
---
The built-in UI must show gross, net and liquidation values side by side for the portfolio's total, each account and each holding, a missing value saying so with the read API's reason, and must mark a holding or lot no longer updated with the removed instances behind it.

## Context

D-77 has tables show the three values side by side, D-78 shows an IKZE's liquidation value as missing until the marginal rate is set, and D-91 marks holdings and lots. Showing the reason next to a missing value is this package's reading. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-77]** Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper: Gross: before any tax or fee; market value, or the accrued value of a bond or deposit. Net: what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees. Liquidation: the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

docs/DESIGN.md#D-77
