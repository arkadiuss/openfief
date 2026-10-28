---
title: Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded.
logic_claims:
  - CLAIM-E7DBDB353368E19A
semantic_clauses:
  - Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4260065e2ed6968e270343fe5b91429657e279f31a29e434f314509c3c9a9af
semantic_inventory:
  - claim_key: CLAIM-E7DBDB353368E19A
    claim_text: Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded
    role: exception
    status: modeled
    span:
      start: 0
      end: 304
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:17.031Z'
id: REQ-ledger-seed-refuse-reseed
type: req
---
Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded.

## Context

D-23 seeds each account once, at go-live, from the broker's current lot-level positions; a second seeding would double the opening lots and cash. The author told the orchestrator, in answer to this task (not yet written into DESIGN.md): opening cash stays a deposit or withdrawal against the equity ledger account, and seeding an account that is already seeded is refused with a clear error. Re-running the identical statement is allowed as a no-op so that a retried upload is harmless. Refines REQ-AUTO-3C2A493C661181C7. Modeled by Claude Code for the M0 seed-guard task; the author has not reviewed it yet.

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md D-23
