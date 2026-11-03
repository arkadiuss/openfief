---
title: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
status: open
text_ref: docs/DESIGN.md#D-162
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-162
  - lane:strict
semantic_text: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
logic_claims:
  - CLAIM-9D4B7C0F489D757B
semantic_clauses:
  - An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e2f693b163e31b2f9452a6bf27dda3991ff452010f6a3e57e915645e317fc25e
semantic_inventory:
  - claim_key: CLAIM-9D4B7C0F489D757B
    claim_text: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it
    role: normative
    status: modeled
    span:
      start: 0
      end: 203
origin:
  kind: agent
  recorded_at: '2026-11-03T11:51:24.277Z'
id: REQ-ingest-payout-setting-history
type: req
---
An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.

## Context

The statement of REQ-ledger-payout-setting-history, now owned by the ingest context, which keeps the payout setting (D-135, D-162). Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> An account's first payout setting covers its whole history; a change takes effect from a day the owner gives.

docs/DESIGN.md D-162

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
