---
title: Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged.
logic_claims:
  - CLAIM-1FC04A0118B07C28
semantic_clauses:
  - Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8a21636d9c88632a80e096b6dac04f5b7ffb291982648dc3f4f03140185a9977
semantic_inventory:
  - claim_key: CLAIM-1FC04A0118B07C28
    claim_text: Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged
    role: normative
    status: modeled
    span:
      start: 0
      end: 206
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:48.732Z'
id: REQ-ledger-app-assertion-flag
type: req
---
Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged.

## Context

Refines the firm requirement REQ-AUTO-D9FBE33256704D04: a mismatch is a flagged error, not a rejected import, so the owner sees it and the rows that caused it stay in the log. Stored assertions are checked again on every import into their account, because a later correction can fix or cause a mismatch. A new report for the same account, date and ledger account replaces the stored one. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
