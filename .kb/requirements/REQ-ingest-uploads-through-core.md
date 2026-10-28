---
title: File uploads and manual bond-lot entries in the built-in UI must go to the core, which hands them to the right connector.
status: open
text_ref: design-decisions:D-49#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-49-3
  - lane:strict
semantic_text: File uploads and manual bond-lot entries in the built-in UI must go to the core, which hands them to the right connector.
logic_claims:
  - CLAIM-9501C4D99C5B4728
semantic_clauses:
  - File uploads and manual bond-lot entries in the built-in UI must go to the core, which hands them to the right connector.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 86165a0ff2ff2a2da532eeb1aa17a61e9d6220fe122cb051ce786432b2062ab7
semantic_inventory:
  - claim_key: CLAIM-9501C4D99C5B4728
    claim_text: File uploads and manual bond-lot entries in the built-in UI must go to the core, which hands them to the right connector
    role: normative
    status: modeled
    span:
      start: 0
      end: 120
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-49#3
  recorded_at: '2026-10-28T05:39:43.753Z'
required_evidence: asserted
id: REQ-ingest-uploads-through-core
type: req
---
File uploads and manual bond-lot entries in the built-in UI must go to the core, which hands them to the right connector.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-49#3
