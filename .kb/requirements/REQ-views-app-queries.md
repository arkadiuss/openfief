---
title: The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account.
status: open
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
semantic_text: The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account.
logic_claims:
  - CLAIM-6B5ED0620547FE66
semantic_clauses:
  - The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 65b7195f49b02392855625f24cd02de9e3e10affc931743c85e1e00d47c9edce
semantic_inventory:
  - claim_key: CLAIM-6B5ED0620547FE66
    claim_text: The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:49.442Z'
id: REQ-views-app-queries
type: req
---
The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account.

## Context

D-6 lists accounts, transactions, tax lots and holdings in the public API, and GR-3 asks for API access to what the owner holds (firm REQ-AUTO-6BEF2198D156F4F0). The views context reads the ledger only through its app queries (context map in docs/dev/code-organization.md), declared here as the consumer-side Ledger port that *ledgerapp.Ledger implements. Telling an unknown account from a failure is what lets the HTTP adapter answer 404 instead of 500. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
