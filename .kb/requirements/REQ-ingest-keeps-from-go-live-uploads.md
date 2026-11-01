---
title: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.'
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.'
logic_claims:
  - CLAIM-CCA301C6324310BF
semantic_clauses:
  - 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bbb7684303df9a33ad3b20fdcdd92c747b69c1fff5fbb58a554758ae9f9d2b4d
semantic_inventory:
  - claim_key: CLAIM-CCA301C6324310BF
    claim_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them'
    role: normative
    status: modeled
    span:
      start: 0
      end: 402
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:33.635Z'
id: REQ-ingest-keeps-from-go-live-uploads
type: req
---
A sync, or a statement uploaded in transactions mode, must not book again what an account's seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.

## Context

D-23 seeds accounts from opening lots with transactions from go-live, and D-136 has later XTB exports continue an account started from its open positions; an export reaches back before go-live, so uploads in transactions mode are kept from go-live like syncs. It replaces the earlier requirement that booked uploads whole. Recorded by the XTB start implementation agent (impl/xtb-start) for D-136; default firmness, not reviewed by the author yet; firm requirements for D-136 are to come from the design session.

## Why this changed

D-136, decided by the author, replaces the reading REQ-ingest-sync-keeps-from-go-live recorded: an XTB account now starts from its open positions too, and later statements continue it from go-live, so the earlier narrow reading no longer holds.

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live. Full history backfill arrives with the tax pack.

docs/DESIGN.md#D-23
