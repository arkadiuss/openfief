---
title: 'The liquidation value must be the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-4
  - lane:strict
text_ref: design-decisions:D-77#4
origin:
  kind: import
  ref: design-decisions:D-77#4
  recorded_at: '2026-10-29T07:59:52.224Z'
semantic_text: 'The liquidation value must be the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.'
logic_claims:
  - CLAIM-B1CCF002C7C52A9C
semantic_clauses:
  - 'The liquidation value must be the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e7bc3218b1160f00db5590563d7221700e65306059e22db40726db62e0a7d0db
semantic_inventory:
  - claim_key: CLAIM-B1CCF002C7C52A9C
    claim_text: 'The liquidation value must be the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions'
    role: normative
    status: modeled
    span:
      start: 0
      end: 205
id: REQ-valuation-liquidation-value
type: req
---
The liquidation value must be the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

## Why this changed

The author's decision D-77 replaces D-25, under which bonds counted in totals at their net-if-redeemed-today value by default, after the early-redemption fee and Belka tax. D-77 generalizes that value to the liquidation value of every holding, account and total, adding early-withdrawal taxes and sale commissions, and makes it one of three values rather than the default: charts start on gross (REQ-ui-value-switch). REQ-AUTO-E10BE8CE13E67657 stated D-25's default, and D-25 has left docs/DESIGN.md, so this requirement replaces it.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values); this is the third. D-78 says how its inputs are estimated: a sale fee rule the owner sets per account (REQ-valuation-sale-fee-rule) and the owner's marginal PIT rate for early IKZE withdrawals (REQ-pl-ikze-early-withdrawal-tax). The implementation's net_if_redeemed bond basis and its fee and tax rules (REQ-valuation-bond-basis, REQ-pl-early-redemption-fee-cap, REQ-pl-belka-tax, REQ-pl-tax-wrapper-exempt) compute part of this value for bonds today. D-25 counted bonds net of tax while stocks stayed gross, so totals mixed two bases (P-40, which D-77 decides). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper:
> - **Gross:** before any tax or fee; market value, or the accrued value of a bond or deposit.
> - **Net:** what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees.
> - **Liquidation:** the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.
>
> Tables show all three side by side. Each chart has a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#4
