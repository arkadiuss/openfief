---
title: The owner must be able to acknowledge an open attention item, and to resolve by hand only a flagged balance assertion's or an import's item; an acknowledged or owner-resolved item must open again when its detail changes.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The owner must be able to acknowledge an open attention item, and to resolve by hand only a flagged balance assertion's or an import's item; an acknowledged or owner-resolved item must open again when its detail changes.
logic_claims:
  - CLAIM-45E0CDBF5D3861CC
semantic_clauses:
  - The owner must be able to acknowledge an open attention item, and to resolve by hand only a flagged balance assertion's or an import's item; an acknowledged or owner-resolved item must open again when its detail changes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 11923d87d065aa8b9cb6d1fa5ea1579f5277acc7ccbb7c3fb517a0c3b1e7908e
semantic_inventory:
  - claim_key: CLAIM-45E0CDBF5D3861CC
    claim_text: The owner must be able to acknowledge an open attention item, and to resolve by hand only a flagged balance assertion's or an import's item; an acknowledged or owner-resolved item must open again when its detail changes
    role: normative
    status: modeled
    span:
      start: 0
      end: 219
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:48.828Z'
id: REQ-attention-owner-actions
type: req
---
The owner must be able to acknowledge an open attention item, and to resolve by hand only a flagged balance assertion's or an import's item; an acknowledged or owner-resolved item must open again when its detail changes.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-store; the price check acknowledgement of D-112 is the model.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
