---
title: Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.
logic_claims:
  - CLAIM-97718BFE48BD8939
semantic_clauses:
  - Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7f6a446f59199ea4deee6352fcfef164d35f5a78455496349e0d2f9055865c4a
semantic_inventory:
  - claim_key: CLAIM-97718BFE48BD8939
    claim_text: Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:21.596Z'
id: REQ-valuation-enqueue-same-transaction
type: req
---
Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.

## Context

D-36 enqueues a revaluation in the same database transaction as the change; the ledger import enqueues through its ImportHook on the import's transaction (ledger/postgres WithTx). Refines REQ-decision-d36-3-revaluation-same-transaction. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date.

docs/DESIGN.md D-36
