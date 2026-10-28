---
title: An import batch must record its institution, ingest path, ingest mode and start date, the number of transactions and balance assertions mapped, the ledger's import summary including failed balance assertions, and every ingest problem.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: An import batch must record its institution, ingest path, ingest mode and start date, the number of transactions and balance assertions mapped, the ledger's import summary including failed balance assertions, and every ingest problem.
logic_claims:
  - CLAIM-D1B6A7A1F24EA12E
semantic_clauses:
  - An import batch must record its institution, ingest path, ingest mode and start date, the number of transactions and balance assertions mapped, the ledger's import summary including failed balance assertions, and every ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ad89d16657f06d4b082934482b3913c7d02ad29c5a91d58123cc7b87089b8443
semantic_inventory:
  - claim_key: CLAIM-D1B6A7A1F24EA12E
    claim_text: An import batch must record its institution, ingest path, ingest mode and start date, the number of transactions and balance assertions mapped, the ledger's import summary including failed balance assertions, and every ingest problem
    role: normative
    status: modeled
    span:
      start: 0
      end: 233
origin:
  kind: agent
  recorded_at: '2026-10-28T06:17:40.099Z'
id: REQ-ingest-batch-record
type: req
---
An import batch must record its institution, ingest path, ingest mode and start date, the number of transactions and balance assertions mapped, the ledger's import summary including failed balance assertions, and every ingest problem.

## Context

D-7 makes the connector's ingest path the only way data enters, and D-18 makes a balance assertion mismatch a flagged error, so each run must leave a record the owner can inspect: what was read, how much was imported or superseded, which assertions failed and which rows were left out and why. Refines REQ-AUTO-D9FBE33256704D04 and REQ-AUTO-41A1F4D687DAA771. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet.

## Source

> Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
