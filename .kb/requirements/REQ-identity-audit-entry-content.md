---
title: Each audit entry must name who, what, which records by ID, when and the outcome.
status: open
tags:
  - strict-modeling
  - lane:strict
  - provenance:design-decisions-d-211-3
text_ref: design-decisions:D-211#3
origin:
  kind: import
  ref: design-decisions:D-211#3
  recorded_at: '2026-11-12T19:18:17.451Z'
semantic_text: Each audit entry must name who, what, which records by ID, when and the outcome.
logic_claims:
  - CLAIM-97787BEFDE66F955
semantic_clauses:
  - Each audit entry must name who, what, which records by ID, when and the outcome.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2f2541d7cfaaa3b7b632cd3ba8b382db0b48d1851923c5cceda3f8d23db30b35
semantic_inventory:
  - claim_key: CLAIM-97787BEFDE66F955
    claim_text: Each audit entry must name who, what, which records by ID, when and the outcome
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
id: REQ-identity-audit-entry-content
type: req
---
Each audit entry must name who, what, which records by ID, when and the outcome.

## Context

The author decided it with D-211. In the design session the author added that entries name people by ID only. The built entry has an actor kind and ID, an action from a closed list, references by kind and ID, a time and an outcome of ok, refused or failed. No end-to-end test proves it, so its evidence floor is asserted.

## Source

> Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12).

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#3