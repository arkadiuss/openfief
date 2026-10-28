---
title: The core must own connector schedules, retries and import batches.
status: open
text_ref: design-decisions:D-49#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-49-2
  - lane:strict
semantic_text: The core must own connector schedules, retries and import batches.
logic_claims:
  - CLAIM-56F139357960D958
semantic_clauses:
  - The core must own connector schedules, retries and import batches.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c5fe7cf1125cfae0e12ffd7410cc19cb4ba86ee68df3a8440640750320f0c9e
semantic_inventory:
  - claim_key: CLAIM-56F139357960D958
    claim_text: The core must own connector schedules, retries and import batches
    role: normative
    status: modeled
    span:
      start: 0
      end: 65
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-49#2
  recorded_at: '2026-10-28T05:39:36.499Z'
required_evidence: asserted
id: REQ-ingest-core-owns-schedules
type: req
---
The core must own connector schedules, retries and import batches.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-49#2
