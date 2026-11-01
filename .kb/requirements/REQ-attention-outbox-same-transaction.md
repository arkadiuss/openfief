---
title: Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.
logic_claims:
  - CLAIM-E661487AD21DB6E1
semantic_clauses:
  - Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ac33d931f1c76687ca653c5c5ef0217e7dc257a2a77d8db7c6d2f96b3f50974
semantic_inventory:
  - claim_key: CLAIM-E661487AD21DB6E1
    claim_text: Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only
    role: normative
    status: modeled
    span:
      start: 0
      end: 162
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:01.854Z'
id: REQ-attention-outbox-same-transaction
type: req
---
Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-outbox-events with the transactional outbox pattern.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
