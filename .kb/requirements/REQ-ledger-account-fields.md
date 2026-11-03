---
title: An account must have an ID, an institution and a name, an account without any of them must be rejected, and the ledger must keep neither the account's sleeve nor its payout setting.
status: open
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
semantic_text: An account must have an ID, an institution and a name, an account without any of them must be rejected, and the ledger must keep neither the account's sleeve nor its payout setting.
logic_claims:
  - CLAIM-5B027F1C2175E5BA
semantic_clauses:
  - An account must have an ID, an institution and a name, an account without any of them must be rejected, and the ledger must keep neither the account's sleeve nor its payout setting.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1854938134c482af365d63c1f22561dc67f6b5e975836c0fc1ac27b44b8dd4ee
semantic_inventory:
  - claim_key: CLAIM-5B027F1C2175E5BA
    claim_text: An account must have an ID, an institution and a name, an account without any of them must be rejected, and the ledger must keep neither the account's sleeve nor its payout setting
    role: normative
    status: modeled
    span:
      start: 0
      end: 180
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:52.153Z'
id: REQ-ledger-account-fields
type: req
---
An account must have an ID, an institution and a name, an account without any of them must be rejected, and the ledger must keep neither the account's sleeve nor its payout setting.

## Context

The ledger keeps what a transaction is booked in. Which sleeve an account is in is the portfolio context's, and the bond payout setting the ingest context's. Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
