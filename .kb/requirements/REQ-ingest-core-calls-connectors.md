---
title: Connectors must serve a Connect service that the core calls to sync and to import files.
status: open
text_ref: design-decisions:D-49#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-49-1
  - lane:strict
semantic_text: Connectors must serve a Connect service that the core calls to sync and to import files.
logic_claims:
  - CLAIM-E4FFA67B9F710F3D
semantic_clauses:
  - Connectors must serve a Connect service that the core calls to sync and to import files.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 738d609f222e963fa65086ff0b3dcd5484eb5c9ada163ccc088f8652db91c5b3
semantic_inventory:
  - claim_key: CLAIM-E4FFA67B9F710F3D
    claim_text: Connectors must serve a Connect service that the core calls to sync and to import files
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-49#1
  recorded_at: '2026-10-28T05:39:29.696Z'
required_evidence: asserted
id: REQ-ingest-core-calls-connectors
type: req
---
Connectors must serve a Connect service that the core calls to sync and to import files.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-49#1
