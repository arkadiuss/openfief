---
title: A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded.
logic_claims:
  - CLAIM-FA06080FF364669F
semantic_clauses:
  - A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b3076f7f78b8e5d7d3cec9a93ae072b8d2a7606f8892dc64a62eb2fe542713f0
semantic_inventory:
  - claim_key: CLAIM-FA06080FF364669F
    claim_text: A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded
    role: normative
    status: modeled
    span:
      start: 0
      end: 353
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:36.589Z'
id: REQ-ingest-scheduler-retry
type: req
---
A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded.

## Context

The connector contract reuses a batch ID on retry, and re-importing is idempotent (D-18), so a retry after a partial import is safe; storing the watermark only on completion means no statement period is skipped. The ledger and valuation inputs are separate database transactions, so the order is ledger first, inputs second, watermark last. Modeled by Claude Code for the core-connectors task (orchestrator brief: the core scheduler, watermarks, connector bindings); the author has not reviewed this requirement yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49

> Re-imports are idempotent, keyed by source plus the institution's own ID.

docs/DESIGN.md D-18
