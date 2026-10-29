---
title: Every completed upload and manual entry batch must be recorded in the import batch history with its source, mode, institution, counts and times, and a batch ID may be recorded only once.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: Every completed upload and manual entry batch must be recorded in the import batch history with its source, mode, institution, counts and times, and a batch ID may be recorded only once.
logic_claims:
  - CLAIM-86C70B5CBF12CA06
semantic_clauses:
  - Every completed upload and manual entry batch must be recorded in the import batch history with its source, mode, institution, counts and times, and a batch ID may be recorded only once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c2410334759908624e0c21d6c667a17aaf0775a738b5fe51dfc65cf8aebb4b3
semantic_inventory:
  - claim_key: CLAIM-86C70B5CBF12CA06
    claim_text: Every completed upload and manual entry batch must be recorded in the import batch history with its source, mode, institution, counts and times, and a batch ID may be recorded only once
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:35.601Z'
id: REQ-ingest-recorded-batches
type: req
---
Every completed upload and manual entry batch must be recorded in the import batch history with its source, mode, institution, counts and times, and a batch ID may be recorded only once.

## Context

D-49: the core owns the import batches. Until now only syncs were recorded; uploads and bond lot entries join them with sources upload and entry. Recording only completed batches is this task's narrow choice. Modeled by Claude Code for the statement upload task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
