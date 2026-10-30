---
title: A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions.
logic_claims:
  - CLAIM-6BCEA07426B522FA
semantic_clauses:
  - A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 24417123cb10b7848b23dccef4e2b20b5575c43c31b646299b95c939673132b2
semantic_inventory:
  - claim_key: CLAIM-6BCEA07426B522FA
    claim_text: A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions
    role: normative
    status: modeled
    span:
      start: 0
      end: 265
origin:
  kind: agent
  recorded_at: '2026-10-29T11:37:01.268Z'
id: REQ-ingest-broker-account-seeding
type: req
---
A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions.

## Context

D-23 seeds accounts from current positions at go-live. A connector usually syncs before the owner maps its broker accounts in the first-run wizard (D-73), and the ledger refuses to seed an account again with changed content, so seeding is tracked per mapped broker account. The mechanism is the implementation's reading. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-23]** Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md#D-23

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
