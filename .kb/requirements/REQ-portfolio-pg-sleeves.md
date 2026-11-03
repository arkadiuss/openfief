---
title: The portfolio's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, refuse a role override that doesn't match a sleeve's role and any change of a stored kind's default role, and keep which sleeve each account is in, refusing a sleeve that is not stored.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The portfolio's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, refuse a role override that doesn't match a sleeve's role and any change of a stored kind's default role, and keep which sleeve each account is in, refusing a sleeve that is not stored.
logic_claims:
  - CLAIM-6080BDABB5FF5886
semantic_clauses:
  - The portfolio's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, refuse a role override that doesn't match a sleeve's role and any change of a stored kind's default role, and keep which sleeve each account is in, refusing a sleeve that is not stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c99de70ab42e551a66804dba997dbba12cfac528f4b99184fbe7e8eb7066715d
semantic_inventory:
  - claim_key: CLAIM-6080BDABB5FF5886
    claim_text: The portfolio's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, refuse a role override that doesn't match a sleeve's role and any change of a stored kind's default role, and keep which sleeve each account is in, refusing a sleeve that is not stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 308
origin:
  kind: agent
  recorded_at: '2026-11-03T11:51:06.528Z'
id: REQ-portfolio-pg-sleeves
type: req
---
The portfolio's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, refuse a role override that doesn't match a sleeve's role and any change of a stored kind's default role, and keep which sleeve each account is in, refusing a sleeve that is not stored.

## Context

Replaces the ledger schema's sleeve tables; nothing is copied from them (D-161). Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
