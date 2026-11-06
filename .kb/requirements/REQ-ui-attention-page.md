---
title: The built-in UI must list attention items on one page by state with their history and a link to where the owner acts, count the open ones in its navigation and overview, and offer the owner's actions to an owner session only.
status: closed
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The built-in UI must list attention items on one page by state with their history and a link to where the owner acts, count the open ones in its navigation and overview, and offer the owner's actions to an owner session only.
logic_claims:
  - CLAIM-E408E81E4A313EC2
semantic_clauses:
  - The built-in UI must list attention items on one page by state with their history and a link to where the owner acts, count the open ones in its navigation and overview, and offer the owner's actions to an owner session only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 435c4e8a4b1ba4076b6cf33adc1246fea0c9f6cedbfda342e7d9ac3237a2f569
semantic_inventory:
  - claim_key: CLAIM-E408E81E4A313EC2
    claim_text: The built-in UI must list attention items on one page by state with their history and a link to where the owner acts, count the open ones in its navigation and overview, and offer the owner's actions to an owner session only
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:38.025Z'
id: REQ-ui-attention-page
type: req
---
The built-in UI must list attention items on one page by state with their history and a link to where the owner acts, count the open ones in its navigation and overview, and offer the owner's actions to an owner session only.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-ui-attention-items-one-place.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130
