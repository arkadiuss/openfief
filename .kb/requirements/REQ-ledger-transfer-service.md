---
title: The core must serve the owner's transfer decisions as TransferService, confirming or rejecting a pair of a withdrawal and a deposit, only for view plugins acting for an owner session.
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: The core must serve the owner's transfer decisions as TransferService, confirming or rejecting a pair of a withdrawal and a deposit, only for view plugins acting for an owner session.
logic_claims:
  - CLAIM-B61CAD9C6EC679A2
semantic_clauses:
  - The core must serve the owner's transfer decisions as TransferService, confirming or rejecting a pair of a withdrawal and a deposit, only for view plugins acting for an owner session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 37002601e9544aa609141662041a935a86320058d24af79f87f628d334b2506c
semantic_inventory:
  - claim_key: CLAIM-B61CAD9C6EC679A2
    claim_text: The core must serve the owner's transfer decisions as TransferService, confirming or rejecting a pair of a withdrawal and a deposit, only for view plugins acting for an owner session
    role: normative
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:39.031Z'
id: REQ-ledger-transfer-service
type: req
---
The core must serve the owner's transfer decisions as TransferService, confirming or rejecting a pair of a withdrawal and a deposit, only for view plugins acting for an owner session.

## Context

D-79 gives each kind of owner entry its own owner-only service; D-105's confirmations are such an entry. Refines REQ-core-owner-entry-services and REQ-ui-transfer-confirmation.

## Source

> What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69).

docs/DESIGN.md#D-79

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
