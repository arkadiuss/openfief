---
title: The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint.
logic_claims:
  - CLAIM-29A26B461CA33E84
semantic_clauses:
  - The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a74504015c16c197917043e5b001893dfb9ebaa650d53716e7baf83edf24b8c7
semantic_inventory:
  - claim_key: CLAIM-29A26B461CA33E84
    claim_text: The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:45.869Z'
id: REQ-ledger-pg-append-only
type: req
---
The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint.

## Context

Refines the firm requirement REQ-AUTO-CACBF41A8E14837D (D-18) down to the database, so a bug or a manual statement cannot rewrite history. Triggers were chosen over REVOKE because in M0 the core connects as the schema owner, which a REVOKE does not bind; the owner can still switch a trigger off, which the fingerprint check on load then detects. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
