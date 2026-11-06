---
title: The ingest store must list completed import batches of every source newest first, by when they finished and then by ID, with their counts, and continue after a given batch.
status: open
text_ref: docs/DESIGN.md#D-193
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-193
  - lane:strict
semantic_text: The ingest store must list completed import batches of every source newest first, by when they finished and then by ID, with their counts, and continue after a given batch.
logic_claims:
  - CLAIM-CA85D56F2D42A05D
semantic_clauses:
  - The ingest store must list completed import batches of every source newest first, by when they finished and then by ID, with their counts, and continue after a given batch.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fabc690c98b7de820b7326cef0a56f93eb2b92a06a9318ba2e2cf7e618367e53
semantic_inventory:
  - claim_key: CLAIM-CA85D56F2D42A05D
    claim_text: The ingest store must list completed import batches of every source newest first, by when they finished and then by ID, with their counts, and continue after a given batch
    role: normative
    status: modeled
    span:
      start: 0
      end: 171
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:15.966Z'
id: REQ-ingest-completed-batches-pages
type: req
---
The ingest store must list completed import batches of every source newest first, by when they finished and then by ID, with their counts, and continue after a given batch.

## Context

D-193 ends the overview with recent activity: imports, syncs, coupons and corrections. The read API's activity reads the imports and syncs from here. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-193]** ... and recent activity: imports, syncs, coupons and corrections.

docs/DESIGN.md#D-193
