---
title: The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored.
logic_claims:
  - CLAIM-78198D70FFA8D894
semantic_clauses:
  - The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4341a7fd3fee979da8f3dcd49b588b1acaef7aa19fc5aca423b7d8b1423c1efd
semantic_inventory:
  - claim_key: CLAIM-78198D70FFA8D894
    claim_text: The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:55.181Z'
id: REQ-ledger-app-queries
type: req
---
The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored.

## Context

Refines the firm requirement REQ-AUTO-BE6EA3B9F864EA40 for the read API (D-6 lists accounts, transactions, tax lots and holdings). The queries load the account's log and call the pure ledger functions; a transaction corrected into another account counts only in its current account. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations.

docs/DESIGN.md D-6
