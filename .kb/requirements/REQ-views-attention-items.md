---
title: The read API must list every attention item with its kind, subject, state, detail, history and whether the owner may resolve it, and how many are open.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The read API must list every attention item with its kind, subject, state, detail, history and whether the owner may resolve it, and how many are open.
logic_claims:
  - CLAIM-1D39BCF19555B52E
semantic_clauses:
  - The read API must list every attention item with its kind, subject, state, detail, history and whether the owner may resolve it, and how many are open.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dca8708784897b5bb0acd691358264175242cd354ad785016da5b5b7dca38401
semantic_inventory:
  - claim_key: CLAIM-1D39BCF19555B52E
    claim_text: The read API must list every attention item with its kind, subject, state, detail, history and whether the owner may resolve it, and how many are open
    role: normative
    status: modeled
    span:
      start: 0
      end: 150
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:30.559Z'
id: REQ-views-attention-items
type: req
---
The read API must list every attention item with its kind, subject, state, detail, history and whether the owner may resolve it, and how many are open.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). The read API is how the UI and scripts read the items (D-6).

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
