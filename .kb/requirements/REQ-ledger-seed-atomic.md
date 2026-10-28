---
title: The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed.
logic_claims:
  - CLAIM-F5D68475971E0AB7
semantic_clauses:
  - The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bd1095d2a2cd4d32e1652132d5c9aa40f1d2f8276181a21c2e12d3e05bbe3b95
semantic_inventory:
  - claim_key: CLAIM-F5D68475971E0AB7
    claim_text: The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed
    role: normative
    status: modeled
    span:
      start: 0
      end: 192
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:26.709Z'
id: REQ-ledger-seed-atomic
type: req
---
The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed.

## Context

A check outside the import transaction, or inside it without a lock, lets two seedings with different keys both see an unseeded account under PostgreSQL's default read committed isolation. The lock is a row lock on the account (FOR NO KEY UPDATE), so ordinary imports that only reference the account are not blocked. The author told the orchestrator, in answer to this task (not yet written into DESIGN.md): opening cash stays a deposit or withdrawal against the equity ledger account, and seeding an account that is already seeded is refused with a clear error. Modeled by Claude Code for the M0 seed-guard task; the author has not reviewed it yet.

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md D-23
