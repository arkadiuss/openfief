---
title: The net value must be what the owner keeps at the planned exit, maturity or a withdrawal from a wrapper once its conditions are met, after the taxes due then and without early-exit fees.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-3
  - lane:strict
text_ref: design-decisions:D-77#3
origin:
  kind: import
  ref: design-decisions:D-77#3
  recorded_at: '2026-10-29T07:58:59.419Z'
semantic_text: The net value must be what the owner keeps at the planned exit, maturity or a withdrawal from a wrapper once its conditions are met, after the taxes due then and without early-exit fees.
logic_claims:
  - CLAIM-2EAE627845A44224
semantic_clauses:
  - The net value must be what the owner keeps at the planned exit, maturity or a withdrawal from a wrapper once its conditions are met, after the taxes due then and without early-exit fees.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4b9b8fe9cd66af78db73cf6d980779d2c663bb98d49b978a98dec974bc1bd62f
semantic_inventory:
  - claim_key: CLAIM-2EAE627845A44224
    claim_text: The net value must be what the owner keeps at the planned exit, maturity or a withdrawal from a wrapper once its conditions are met, after the taxes due then and without early-exit fees
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
id: REQ-valuation-net-value
type: req
---
The net value must be what the owner keeps at the planned exit, maturity or a withdrawal from a wrapper once its conditions are met, after the taxes due then and without early-exit fees.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values); this is the second. The tax wrapper comes from D-76 (REQ-ledger-account-tax-wrapper), so the taxes due at exit depend on it, while early-exit fees belong to the liquidation value (REQ-valuation-liquidation-value). The implementation's net_if_redeemed bond basis (REQ-valuation-bond-basis) deducts the early-redemption fee, so it is not this value. D-77 replaces D-25's single configurable bond basis, which counted bonds net of tax while stocks stayed gross, so totals mixed two bases (P-40, which D-77 decides). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#3
