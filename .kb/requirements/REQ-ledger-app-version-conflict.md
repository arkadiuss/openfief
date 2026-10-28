---
title: 'Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts.'
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: 'Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts.'
logic_claims:
  - CLAIM-7F98397E81D0B356
semantic_clauses:
  - 'Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d9ff8efccf7fc2f11df6df9f8601971a26e3e1d12dc1f4b15f5cd3e41ae34241
semantic_inventory:
  - claim_key: CLAIM-7F98397E81D0B356
    claim_text: 'Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts'
    role: normative
    status: modeled
    span:
      start: 0
      end: 263
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:43.151Z'
id: REQ-ledger-app-version-conflict
type: req
---
Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts.

## Context

The orchestrator asked for concurrency safety: an upload and a watched-folder import of the same export can run at once. Without it two version 2 rows could exist and the log would be invalid (ledger.ErrInvalidLog). The store enforces uniqueness of source, external ID and version; the use case retries up to MaxAppendAttempts (3), which always suffices when the competitor appended the same content. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
