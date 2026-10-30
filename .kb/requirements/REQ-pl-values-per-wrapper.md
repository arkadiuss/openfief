---
title: 'The Polish pack must value a holding priced at market, a bond and cash in each tax wrapper: taxable with Belka tax on the gain at the exit and on the proceeds'' gain after the sale cost if liquidated, IKE untaxed at the exit and taxed like a taxable sale on an early return, IKZE at 10% of the amount at the exit and at the owner''s marginal rate before it, and OKI with no Belka tax, each tax rounded half up to grosze.'
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: 'The Polish pack must value a holding priced at market, a bond and cash in each tax wrapper: taxable with Belka tax on the gain at the exit and on the proceeds'' gain after the sale cost if liquidated, IKE untaxed at the exit and taxed like a taxable sale on an early return, IKZE at 10% of the amount at the exit and at the owner''s marginal rate before it, and OKI with no Belka tax, each tax rounded half up to grosze.'
logic_claims:
  - CLAIM-50CF2E6492F8553E
semantic_clauses:
  - 'The Polish pack must value a holding priced at market, a bond and cash in each tax wrapper: taxable with Belka tax on the gain at the exit and on the proceeds'' gain after the sale cost if liquidated, IKE untaxed at the exit and taxed like a taxable sale on an early return, IKZE at 10% of the amount at the exit and at the owner''s marginal rate before it, and OKI with no Belka tax, each tax rounded half up to grosze.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9d9f907ca63c0fc35041ba46f595d067f26b15d478e1c2142b9e348a541d7afb
semantic_inventory:
  - claim_key: CLAIM-50CF2E6492F8553E
    claim_text: 'The Polish pack must value a holding priced at market, a bond and cash in each tax wrapper: taxable with Belka tax on the gain at the exit and on the proceeds'' gain after the sale cost if liquidated, IKE untaxed at the exit and taxed like a taxable sale on an early return, IKZE at 10% of the amount at the exit and at the owner''s marginal rate before it, and OKI with no Belka tax, each tax rounded half up to grosze'
    role: normative
    status: modeled
    span:
      start: 0
      end: 417
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:16.196Z'
id: REQ-pl-values-per-wrapper
type: req
---
The Polish pack must value a holding priced at market, a bond and cash in each tax wrapper: taxable with Belka tax on the gain at the exit and on the proceeds' gain after the sale cost if liquidated, IKE untaxed at the exit and taxed like a taxable sale on an early return, IKZE at 10% of the amount at the exit and at the owner's marginal rate before it, and OKI with no Belka tax, each tax rounded half up to grosze.

## Context

D-77 has the pack compute the three values per product kind and wrapper; D-76 names the wrappers, D-78 the IKZE rate. The rules per wrapper are this package's narrowest reading of Polish tax law: IKE's planned exit untaxed (art. 21 ust. 1 pkt 58 of the PIT act), an early return taxed at 19% on its gain; IKZE's exit at 10% of the amount (art. 30 ust. 1 pkt 13); per-holding gains with no offsetting of losses across holdings; rounding half up as D-37 does for Belka tax. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-pl-belka-tax: D-77 replaced the net-if-redeemed value with net and liquidation values per wrapper. It replaces REQ-pl-tax-wrapper-exempt: An early return from IKE is taxed on its gain in the liquidation value (D-77). The superseded requirement is closed.

## Source

> **[D-77]** Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper: Gross: before any tax or fee; market value, or the accrued value of a bond or deposit. Net: what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees. Liquidation: the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

docs/DESIGN.md#D-77
