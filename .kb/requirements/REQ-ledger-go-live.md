---
title: An account's go-live date must be the date of its latest current seeding transaction, an opening lot or opening cash, and an account with none must count as not seeded.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: An account's go-live date must be the date of its latest current seeding transaction, an opening lot or opening cash, and an account with none must count as not seeded.
logic_claims:
  - CLAIM-EFFF0296F079F43B
semantic_clauses:
  - An account's go-live date must be the date of its latest current seeding transaction, an opening lot or opening cash, and an account with none must count as not seeded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 15861049836ad594b7339738a24a418ce54898a8b946de8c25f6bc5129bd3a93
semantic_inventory:
  - claim_key: CLAIM-EFFF0296F079F43B
    claim_text: An account's go-live date must be the date of its latest current seeding transaction, an opening lot or opening cash, and an account with none must count as not seeded
    role: normative
    status: modeled
    span:
      start: 0
      end: 167
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:52.012Z'
id: REQ-ledger-go-live
type: req
---
An account's go-live date must be the date of its latest current seeding transaction, an opening lot or opening cash, and an account with none must count as not seeded.

## Context

The ingest context asks the ledger which accounts are seeded, and from when, to choose a sync's mode (D-95) and to leave out transactions before go-live (D-23). Modeled by Claude Code for the market-data work package (D-93, D-94, D-95, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-23]** Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md#D-23
