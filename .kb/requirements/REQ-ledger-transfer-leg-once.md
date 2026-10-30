---
title: Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race.
logic_claims:
  - CLAIM-15A99C908FFCB459
semantic_clauses:
  - Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 34514516c3748089c57dd686025bd367da80a1df1f0831753abe5d3e424a2fbc
semantic_inventory:
  - claim_key: CLAIM-15A99C908FFCB459
    claim_text: Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race
    role: normative
    status: modeled
    span:
      start: 0
      end: 153
origin:
  kind: agent
  recorded_at: '2026-10-30T06:06:11.436Z'
id: REQ-ledger-transfer-leg-once
type: req
---
Confirming a transfer must be refused when its withdrawal or its deposit is already a leg of another confirmed transfer, even when two confirmations race.

## Context

One deposit can't be the other half of two withdrawals: the money would leave the time-weighted return's cash flows twice. D-105 suggests the nearest unmatched deposit; this keeps confirmed transfers matched one to one. This context's reading of D-105; reason given in the design notes, not by the author.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
