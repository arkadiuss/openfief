---
title: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them.'
status: open
text_ref: docs/DESIGN.md#D-140
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-140
  - lane:strict
semantic_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them.'
logic_claims:
  - CLAIM-0288B79F8E43FFDB
semantic_clauses:
  - 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 00f74149b855cba0109d36a84867eddc90cc6a963d2694d7032fea607c666ccb
semantic_inventory:
  - claim_key: CLAIM-0288B79F8E43FFDB
    claim_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them'
    role: normative
    status: modeled
    span:
      start: 0
      end: 529
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:56.567Z'
id: REQ-ingest-keeps-from-go-live-opening-uploads
type: req
---
A sync, or a statement uploaded in transactions mode, must not book again what an account's seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them.

## Why this changed

On 2026-11-01 the author recorded D-140: an opening upload of a statement covering several accounts seeds only the accounts not seeded yet. REQ-ingest-keeps-from-go-live-uploads had opening uploads booked as the owner sent them, so a statement of linked accounts whose other account had started was refused whole by the ledger; this requirement replaces it.

## Context

D-23 seeds an account once, D-95 and D-136 continue it from its go-live, and D-160 records an opening upload's go-live at once. An account started from its whole history (D-136) counts as started though it has no go-live. The account the statement was uploaded for stays whole, so an owner's opening override for a seeded account is still refused by the ledger rather than skipped.

## Source

> **[D-140]** An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

docs/DESIGN.md#D-140
