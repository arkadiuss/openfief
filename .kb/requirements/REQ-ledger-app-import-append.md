---
title: Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version.
logic_claims:
  - CLAIM-B04DE71EC7253A00
semantic_clauses:
  - Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fa22447243a45c4445263e54b7b9a4ed912646ac1d484ae2e872db225766c948
semantic_inventory:
  - claim_key: CLAIM-B04DE71EC7253A00
    claim_text: Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version
    role: normative
    status: modeled
    span:
      start: 0
      end: 241
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:37.783Z'
id: REQ-ledger-app-import-append
type: req
---
Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version.

## Context

Refines the firm requirements REQ-AUTO-CACBF41A8E14837D, REQ-AUTO-F19F621B67F14F4F and REQ-AUTO-22E70B6530E994B1 (D-18) for the import use case that the ingest context calls. The pure decision is ledger.DecideImport; this requirement is about loading every stored version first and appending only what it decides, so a re-import of a whole broker export is a no-op. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
