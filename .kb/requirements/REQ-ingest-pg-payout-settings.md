---
title: The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.
status: open
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
semantic_text: The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.
logic_claims:
  - CLAIM-AB1B1556AB5E5046
semantic_clauses:
  - The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 36f6dcf94fef1f9f809fea26db6a5942ddb7ed4ced60fe7905abcd4586bd3317
semantic_inventory:
  - claim_key: CLAIM-AB1B1556AB5E5046
    claim_text: The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account
    role: normative
    status: modeled
    span:
      start: 0
      end: 277
origin:
  kind: agent
  recorded_at: '2026-11-03T11:51:38.608Z'
id: REQ-ingest-pg-payout-settings
type: req
---
The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.

## Context

The ledger columns that held it are dropped with nothing copied (D-161). Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
