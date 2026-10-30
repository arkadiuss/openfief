---
title: A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded.
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded.
logic_claims:
  - CLAIM-63D0850AA2500FBA
semantic_clauses:
  - A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c6f821b79fa63d3bb29f1caf366bb3debd992bc1171f51c99e12c849123f7117
semantic_inventory:
  - claim_key: CLAIM-63D0850AA2500FBA
    claim_text: A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded
    role: normative
    status: modeled
    span:
      start: 0
      end: 311
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:01.601Z'
id: REQ-valuation-bond-holding-values
type: req
---
A holding of a bond with stored series terms must be valued as the sum of its open lots' gross accrued values, each lot bought on its open date at its cost per bond, with net and liquidation values summed from what the pack computes per lot in the account's tax wrapper, the pack's default when none is recorded.

## Context

Supersedes REQ-valuation-bond-holding, which counted one configured basis that D-77 removed. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-valuation-bond-holding: D-77 removed the configured bond basis. It replaces REQ-valuation-bond-basis: D-77 removed the configured bond basis; gross, net and liquidation replace it. The superseded requirement is closed.

## Source

> **[D-77]** Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper: Gross: before any tax or fee; market value, or the accrued value of a bond or deposit. Net: what the owner keeps at the planned exit (maturity, or a withdrawal from a wrapper once its conditions are met), after the taxes due then and without early-exit fees. Liquidation: the cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions.

docs/DESIGN.md#D-77
