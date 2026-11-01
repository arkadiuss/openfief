---
title: 'A check''s findings must change attention items by fixed rules: one item per kind and subject, raised when found, opened again when its detail changes, and resolved when no longer found.'
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: 'A check''s findings must change attention items by fixed rules: one item per kind and subject, raised when found, opened again when its detail changes, and resolved when no longer found.'
logic_claims:
  - CLAIM-60177075F6AC2CF6
semantic_clauses:
  - 'A check''s findings must change attention items by fixed rules: one item per kind and subject, raised when found, opened again when its detail changes, and resolved when no longer found.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 73bfb6c4e624ac652811b773c0c595b8dcabb3ec6605fc5440dd02b815cd4868
semantic_inventory:
  - claim_key: CLAIM-60177075F6AC2CF6
    claim_text: 'A check''s findings must change attention items by fixed rules: one item per kind and subject, raised when found, opened again when its detail changes, and resolved when no longer found'
    role: normative
    status: modeled
    span:
      start: 0
      end: 184
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:43.125Z'
id: REQ-attention-reconcile
type: req
---
A check's findings must change attention items by fixed rules: one item per kind and subject, raised when found, opened again when its detail changes, and resolved when no longer found.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-store.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
