---
title: An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts.
logic_claims:
  - CLAIM-98ABCCC6310C9B66
semantic_clauses:
  - An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5bf900f71e15d97e22e096632c53648c677d0087443d04577972d7c692aa00f5
semantic_inventory:
  - claim_key: CLAIM-98ABCCC6310C9B66
    claim_text: An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 129
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:06.072Z'
id: REQ-ledger-app-accounts
type: req
---
An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts.

## Context

Refines REQ-ledger-app-registry without the sleeve check, which moved to the portfolio context's registry. Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
