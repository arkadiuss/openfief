---
title: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state.
logic_claims:
  - CLAIM-7239C60159E2987E
semantic_clauses:
  - The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d66cf6a3005b795e95f244c838a64c3c15e73824500d691f2c14b167dcb6f153
semantic_inventory:
  - claim_key: CLAIM-7239C60159E2987E
    claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state
    role: normative
    status: modeled
    span:
      start: 0
      end: 253
origin:
  kind: agent
  recorded_at: '2026-10-29T17:20:16.524Z'
id: REQ-ingest-scheduler-when-due
type: req
---
The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state.

## Context

D-49 makes the core own connectors' schedules. Modeled by Claude Code for the market-data work package (impl/market-data); the author has not reviewed it yet.

## Why this changed

It replaces REQ-ingest-scheduler-due without that requirement's mode clause, opening until one opening batch completes and transactions after it, because D-95 makes a seeded account sync in transactions mode on its own: the mode now follows the ledger's seeding of the bound accounts, as REQ-ingest-sync-mode-from-ledger states.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md#D-49

> **[D-95]** When an account has already been seeded from opening lots, the core syncs it in transactions mode on its own; no manual step switches it.

docs/DESIGN.md#D-95
