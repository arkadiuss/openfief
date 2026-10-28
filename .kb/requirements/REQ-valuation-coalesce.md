---
title: Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates.
logic_claims:
  - CLAIM-D9E6D69119281112
semantic_clauses:
  - Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6facea25c39b75eb2e7ab06bc57dbd5041f316772b689bfd75d596e797efd584
semantic_inventory:
  - claim_key: CLAIM-D9E6D69119281112
    claim_text: Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:45.076Z'
id: REQ-valuation-coalesce
type: req
---
Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates.

## Context

D-36 coalesces pending revaluations per account from the earliest affected date, so a burst of imports and price updates costs one revaluation per account. Refines REQ-decision-d36-5-revaluation-coalescing. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date.

docs/DESIGN.md D-36
