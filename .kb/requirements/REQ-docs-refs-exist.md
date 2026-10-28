---
title: Every kibi requirement, scenario, test or symbol ID and every task name that a docs page mentions must exist.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: Every kibi requirement, scenario, test or symbol ID and every task name that a docs page mentions must exist.
logic_claims:
  - CLAIM-0083CE9511B4CC2A
semantic_clauses:
  - Every kibi requirement, scenario, test or symbol ID and every task name that a docs page mentions must exist.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e7eb5c07ce60c4fe1a9804e68aa89435d02d07a30d50a9ed2d45101bd3953737
semantic_inventory:
  - claim_key: CLAIM-0083CE9511B4CC2A
    claim_text: Every kibi requirement, scenario, test or symbol ID and every task name that a docs page mentions must exist
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:33:00.385Z'
rationale: Pages cite requirements and tasks by name; a renamed or deleted one would otherwise leave a stale reference.
id: REQ-docs-refs-exist
type: req
---
Every kibi requirement, scenario, test or symbol ID and every task name that a docs page mentions must exist.

## Context

Pages cite requirements and tasks by name; a renamed or deleted one would otherwise leave a stale reference. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
