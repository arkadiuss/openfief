---
title: The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail.
logic_claims:
  - CLAIM-9FF3D9D9E3634B7A
semantic_clauses:
  - The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3f82a8a0dceeb16b17d52ad85b175ed19f00851e7809f9a798738c5d41185fa4
semantic_inventory:
  - claim_key: CLAIM-9FF3D9D9E3634B7A
    claim_text: The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:07.862Z'
id: REQ-attention-relay
type: req
---
The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-outbox-events; GR-12 keeps account numbers and amounts out of logs.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
