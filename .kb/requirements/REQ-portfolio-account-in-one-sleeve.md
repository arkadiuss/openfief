---
title: The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve.
status: open
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
semantic_text: The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve.
logic_claims:
  - CLAIM-3D9E32EA7A9495C3
semantic_clauses:
  - The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ba37eca95ee46f6d6b0974d06b6287b01f2cc77755dbbc7999fb386d30e5ea82
semantic_inventory:
  - claim_key: CLAIM-3D9E32EA7A9495C3
    claim_text: The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve
    role: normative
    status: modeled
    span:
      start: 0
      end: 228
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:36.370Z'
id: REQ-portfolio-account-in-one-sleeve
type: req
---
The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve.

## Context

Every account belongs to one sleeve (glossary); the portfolio context keeps that membership in its own schema and reads accounts from the ledger through its app port. Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
