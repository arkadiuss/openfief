---
title: Every instrument the read API names, and the instrument detail, must carry its display name when it has one.
status: open
text_ref: docs/DESIGN.md#D-197
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-197
  - lane:strict
semantic_text: Every instrument the read API names, and the instrument detail, must carry its display name when it has one.
logic_claims:
  - CLAIM-7D197CAB8E47EBCF
semantic_clauses:
  - Every instrument the read API names, and the instrument detail, must carry its display name when it has one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9b520bf9533c23d2575b8ae0f82b61febcd6d84c61d08fdaa61c805c2b9b2341
semantic_inventory:
  - claim_key: CLAIM-7D197CAB8E47EBCF
    claim_text: Every instrument the read API names, and the instrument detail, must carry its display name when it has one
    role: normative
    status: modeled
    span:
      start: 0
      end: 107
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:26.859Z'
id: REQ-views-instrument-names
type: req
---
Every instrument the read API names, and the instrument detail, must carry its display name when it has one.

## Context

Refines D-197 for the read API, so every view and script shows the same name. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-197]** An instrument has a display name: the first one a connector delivers for it, such as XTB's display name, IBKR's description or the market-data lookup's, which the owner can change. Pages show it before the ticker and ISIN.

docs/DESIGN.md#D-197
