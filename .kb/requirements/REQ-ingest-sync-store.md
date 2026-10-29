---
title: The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.
logic_claims:
  - CLAIM-E54E7CB819CE27A7
semantic_clauses:
  - The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ebe3251247a1ae0e13783116d42b4965c3b6df14e848deee08f6154d812eb0bf
semantic_inventory:
  - claim_key: CLAIM-E54E7CB819CE27A7
    claim_text: The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only
    role: normative
    status: modeled
    span:
      start: 0
      end: 296
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:48.354Z'
id: REQ-ingest-sync-store
type: req
---
The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.

## Context

D-14 puts state in PostgreSQL and each context owns its schema (docs/dev/code-organization.md); the orchestrator offered a connectors schema or ingest's own, and the ingest schema follows the one-schema-per-context rule. GR-12 keeps amounts and account numbers out. Modeled by Claude Code for the core-connectors task (orchestrator brief: the core scheduler, watermarks, connector bindings); the author has not reviewed this requirement yet.

## Source

> PostgreSQL via CloudNativePG, and no NATS in M0.

docs/DESIGN.md D-14
