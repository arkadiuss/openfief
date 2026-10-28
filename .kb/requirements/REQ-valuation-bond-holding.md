---
title: A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise.
logic_claims:
  - CLAIM-97670979153E46CE
semantic_clauses:
  - A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 96664829bfa8b6891eeb3753358a1550f40ddc7c56b68044b841023ba5ae3eb6
semantic_inventory:
  - claim_key: CLAIM-97670979153E46CE
    claim_text: A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise
    role: exception
    status: modeled
    span:
      start: 0
      end: 285
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:03.592Z'
id: REQ-valuation-bond-holding
type: req
---
A holding of an instrument whose ISIN has stored bond series terms must be valued as the sum of its open lots' statutory values on the configured bond basis, each lot bought on its open date at its cost per bond, in the account's tax wrapper, which is taxable unless recorded otherwise.

## Context

Treasury bonds enter as manual lots (D-9) and are valued by the issue letters' rules in internal/jurisdiction/pl; which value counts in totals is the configured basis, net if redeemed by default (D-25). Polish retail bonds have an ISIN per series, so terms are keyed by ISIN. Refines REQ-AUTO-0C90AD924924F4F6 and REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Which bond value counts in totals is configurable; the default is net-if-redeemed-today

docs/DESIGN.md D-25
