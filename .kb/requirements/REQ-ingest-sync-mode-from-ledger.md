---
title: 'The core must choose a connector instance''s sync batches from the ledger''s seeding of the accounts bound to it: one transactions batch when every bound account is seeded or none is bound, one opening batch when none is seeded, and an opening batch then a transactions batch, both from the watermark the instance had before them, when only some are.'
status: open
text_ref: docs/DESIGN.md#D-95
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-95
  - lane:strict
semantic_text: 'The core must choose a connector instance''s sync batches from the ledger''s seeding of the accounts bound to it: one transactions batch when every bound account is seeded or none is bound, one opening batch when none is seeded, and an opening batch then a transactions batch, both from the watermark the instance had before them, when only some are.'
logic_claims:
  - CLAIM-A39F854D75E95413
semantic_clauses:
  - 'The core must choose a connector instance''s sync batches from the ledger''s seeding of the accounts bound to it: one transactions batch when every bound account is seeded or none is bound, one opening batch when none is seeded, and an opening batch then a transactions batch, both from the watermark the instance had before them, when only some are.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e3e5102958a9c7ae79c36323411849ad471b3dcf65e696a44072a8a8770814f3
semantic_inventory:
  - claim_key: CLAIM-A39F854D75E95413
    claim_text: 'The core must choose a connector instance''s sync batches from the ledger''s seeding of the accounts bound to it: one transactions batch when every bound account is seeded or none is bound, one opening batch when none is seeded, and an opening batch then a transactions batch, both from the watermark the instance had before them, when only some are'
    role: normative
    status: modeled
    span:
      start: 0
      end: 347
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:31.159Z'
id: REQ-ingest-sync-mode-from-ledger
type: req
---
The core must choose a connector instance's sync batches from the ledger's seeding of the accounts bound to it: one transactions batch when every bound account is seeded or none is bound, one opening batch when none is seeded, and an opening batch then a transactions batch, both from the watermark the instance had before them, when only some are.

## Context

D-95 has the core sync an account already seeded from opening lots in transactions mode on its own. Earlier the instance kept a seeded flag of its own, so an account seeded by an upload, or a seeded account sharing an instance with a newly mapped one (D-72), was synced in opening mode. Reading the seeding from the ledger makes the mode follow the accounts themselves. Modeled by Claude Code for the market-data work package (D-93, D-94, D-95, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-95]** When an account has already been seeded from opening lots, the core syncs it in transactions mode on its own; no manual step switches it.

docs/DESIGN.md#D-95

> **[D-23]** Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md#D-23
