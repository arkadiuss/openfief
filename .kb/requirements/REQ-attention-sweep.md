---
title: The core must run every attention check, each kind's reconciliation in its own database transaction, with a failing check not stopping the others, and must let a check raise and clear a single item.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The core must run every attention check, each kind's reconciliation in its own database transaction, with a failing check not stopping the others, and must let a check raise and clear a single item.
logic_claims:
  - CLAIM-20C69A73462C7214
semantic_clauses:
  - The core must run every attention check, each kind's reconciliation in its own database transaction, with a failing check not stopping the others, and must let a check raise and clear a single item.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 879b624ef63aa3524f54f96980f424cb71ec2c2c7b38b8efa784c9917d2eff8f
semantic_inventory:
  - claim_key: CLAIM-20C69A73462C7214
    claim_text: The core must run every attention check, each kind's reconciliation in its own database transaction, with a failing check not stopping the others, and must let a check raise and clear a single item
    role: normative
    status: modeled
    span:
      start: 0
      end: 197
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:54.147Z'
id: REQ-attention-sweep
type: req
---
The core must run every attention check, each kind's reconciliation in its own database transaction, with a failing check not stopping the others, and must let a check raise and clear a single item.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-checks.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
