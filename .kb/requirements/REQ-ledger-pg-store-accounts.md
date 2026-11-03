---
title: The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
logic_claims:
  - CLAIM-EDF3153345BEC045
semantic_clauses:
  - The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3fe8ce43f5dabcb31cf1c9f26c2346819117f3aee9802fb28b1b90445ea52c6d
semantic_inventory:
  - claim_key: CLAIM-EDF3153345BEC045
    claim_text: The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:21.044Z'
id: REQ-ledger-pg-store-accounts
type: req
---
The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.

## Context

The ledger schema no longer holds sleeves, sleeve kinds or the payout setting; migration 0008 drops them with nothing copied. Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
