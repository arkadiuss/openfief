---
title: The portfolio context must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: The portfolio context must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
logic_claims:
  - CLAIM-5F7438ED85C46D67
semantic_clauses:
  - The portfolio context must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64fdef3ac0436b7888cea97e5ce3a97068734640180f529a8877da721965521b
semantic_inventory:
  - claim_key: CLAIM-5F7438ED85C46D67
    claim_text: The portfolio context must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role
    role: normative
    status: modeled
    span:
      start: 0
      end: 272
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:50.826Z'
id: REQ-portfolio-registry-checks-sleeves
type: req
---
The portfolio context must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.

## Context

The same rules REQ-ledger-app-sleeve-kinds put on the ledger, now run by the portfolio context, which stores the sleeves. Modeled by Claude Code for DDD-1; the author has not reviewed the wording.

## Why this changed

DDD-1, author decision 2026-11-03: the ledger knows nothing of sleeves or of the bond payout setting. docs/dev/code-organization.md's context map says portfolio owns sleeves and which account is in which sleeve, and ingest owns the account settings connectors read, such as the bond payout setting (D-135). Nothing stored before the first release is migrated (D-161): the ledger columns and tables are dropped and the demo is reseeded.

## Source

> `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves.

> `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135)

docs/dev/code-organization.md#bounded-contexts
