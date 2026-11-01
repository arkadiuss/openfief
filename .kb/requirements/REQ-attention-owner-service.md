---
title: The core must serve the AttentionService on its plugin port to view plugins only, for an owner session only, so the owner acknowledges and resolves attention items.
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: The core must serve the AttentionService on its plugin port to view plugins only, for an owner session only, so the owner acknowledges and resolves attention items.
logic_claims:
  - CLAIM-8C30C655BF772057
semantic_clauses:
  - The core must serve the AttentionService on its plugin port to view plugins only, for an owner session only, so the owner acknowledges and resolves attention items.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f86eee66c5aadf7cd2e78736ed89cebac0def479f8203a4fecb34260be2202a5
semantic_inventory:
  - claim_key: CLAIM-8C30C655BF772057
    claim_text: The core must serve the AttentionService on its plugin port to view plugins only, for an owner session only, so the owner acknowledges and resolves attention items
    role: normative
    status: modeled
    span:
      start: 0
      end: 163
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:23.083Z'
id: REQ-attention-owner-service
type: req
---
The core must serve the AttentionService on its plugin port to view plugins only, for an owner session only, so the owner acknowledges and resolves attention items.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Follows D-79 and D-115 for this kind of entry.

## Source

> What the owner enters by hand reaches the core through one service per kind of entry [...] Each is owner-only.

docs/DESIGN.md#D-79
