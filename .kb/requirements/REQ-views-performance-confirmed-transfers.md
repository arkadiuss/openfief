---
title: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows.
logic_claims:
  - CLAIM-E94EA8621E2FA705
semantic_clauses:
  - The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: aa5eeac6ec0f25b816347174771368f7d0b76423e1661e09f45b025f6dbe1906
semantic_inventory:
  - claim_key: CLAIM-E94EA8621E2FA705
    claim_text: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:21.454Z'
id: REQ-views-performance-confirmed-transfers
type: req
---
The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows.

## Context

D-61 says transfers between the owner's own accounts are not cash flows; D-105 makes that depend on the owner's confirmation. Refines the firm REQ-portfolio-twr-confirmed-transfers-only for the read API's series, keeping the in-transit rule built before.

## Why this changed

It replaces REQ-views-performance-transfers: D-105 leaves out of the cash flows only the transfers the owner confirmed, not every pair the matching rule finds. The superseded requirement is closed.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105

> Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows.

docs/DESIGN.md#D-61
