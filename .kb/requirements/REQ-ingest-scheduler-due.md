---
title: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.
status: closed
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.
logic_claims:
  - CLAIM-ED7D7D8E5EF86F7D
semantic_clauses:
  - The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c65543167b9e8d9452495545f31b46b67c6a72b8c4e4cd9ed4e1e390f1b63040
semantic_inventory:
  - claim_key: CLAIM-ED7D7D8E5EF86F7D
    claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state
    role: normative
    status: modeled
    span:
      start: 0
      end: 331
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:31.580Z'
id: REQ-ingest-scheduler-due
type: req
---
The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.

## Context

Closed: superseded by REQ-ingest-scheduler-when-due. D-95 makes a seeded account sync in transactions mode on its own, so the mode follows the ledger. Modeled by Claude Code for the market-data work package (impl/market-data); the author has not reviewed it yet.
