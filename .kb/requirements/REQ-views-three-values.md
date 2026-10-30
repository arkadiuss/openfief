---
title: The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason.
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason.
logic_claims:
  - CLAIM-7DF79E9014C7A86B
semantic_clauses:
  - The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7716cabc4a85cbc7b279cc6b51537566d05e4c732abd6cc2cad33d2969e0e9fc
semantic_inventory:
  - claim_key: CLAIM-7DF79E9014C7A86B
    claim_text: The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason
    role: normative
    status: modeled
    span:
      start: 0
      end: 288
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:06.737Z'
id: REQ-views-three-values
type: req
---
The read API must serve each stored valuation with its gross, net and liquidation values or the reason each is missing, totals with the three values left unset when any part lacks one, and the allocation of the value kind asked for, gross by default, listing what lacks it with the reason.

## Context

D-77 gives every holding, account and total three values, gross by default in charts; D-36 stores them. Leaving a total unset rather than summing what has a value is this package's reading of D-78's 'shown as missing'. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-77]** Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper: Gross: before any tax or fee; market value, or the accrued value of a bond or deposit. Net: what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees. Liquidation: the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

docs/DESIGN.md#D-77
