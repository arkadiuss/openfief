---
title: 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.'
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.'
logic_claims:
  - CLAIM-9AA347F7A866EFE8
semantic_clauses:
  - 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cc8a32d6140e2f68da3442eca49da23d6350e76ec92d64f67da1cfbf9041a96e
semantic_inventory:
  - claim_key: CLAIM-9AA347F7A866EFE8
    claim_text: 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them'
    role: normative
    status: modeled
    span:
      start: 0
      end: 328
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:41.644Z'
id: REQ-ingest-sync-keeps-from-go-live
type: req
---
A sync must not book again what an account's seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.

## Context

D-23 imports transactions from go-live on, and D-95's opening-then-transactions sync would otherwise book a newly seeded account's earlier transactions on top of its opening lots. Uploads and bond lots stay as they were: the owner chooses what they hold (D-109, D-98). Modeled by Claude Code for the market-data work package (D-93, D-94, D-95, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-23]** Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md#D-23

> **[D-95]** When an account has already been seeded from opening lots, the core syncs it in transactions mode on its own; no manual step switches it.

docs/DESIGN.md#D-95
