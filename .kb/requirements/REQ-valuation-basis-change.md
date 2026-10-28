---
title: Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis.
logic_claims:
  - CLAIM-7CC9DCDD33B4D25A
semantic_clauses:
  - Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60706a52f63b9bd7fca0919ba585dd0318158de982f2513dba3d7c132b58f6e7
semantic_inventory:
  - claim_key: CLAIM-7CC9DCDD33B4D25A
    claim_text: Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:57.191Z'
id: REQ-valuation-basis-change
type: req
---
Applying the configured bond basis must revalue every account from its first transaction when it differs from the recorded basis, must not revalue anything when it is the same, and must reject the market basis.

## Context

D-36 makes the bond basis configurable and says changing it revalues everything; comparing with the recorded basis keeps an unchanged restart from revaluing. Refines REQ-decision-d36-9-basis-change-revalues-all. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything.

docs/DESIGN.md D-36
