---
title: The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
logic_claims:
  - CLAIM-2EA08200020CBC78
semantic_clauses:
  - The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9282eb96b37667eed88350ccf78dc22976e4e8827db4337823509e7447878639
semantic_inventory:
  - claim_key: CLAIM-2EA08200020CBC78
    claim_text: The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:35.283Z'
id: REQ-ledger-pg-store
type: req
---
The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.

## Context

D-14 picks PostgreSQL (CloudNativePG). code-organization.md proposes that each context owns its schema (P), so the tables live in the ledger schema with their own migration table, which also lets River's migrations coexist (D-36). Money is never float (AGENTS.md, D-20), so numeric columns are written and read as decimal text. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> PostgreSQL via CloudNativePG, and no NATS in M0.

docs/DESIGN.md D-14
