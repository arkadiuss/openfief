---
title: The built-in UI must list the suggested, confirmed and rejected transfers on a transfers page where an owner session confirms or rejects each suggestion through the core's TransferService and can correct either decision later, while a read session only sees them.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: The built-in UI must list the suggested, confirmed and rejected transfers on a transfers page where an owner session confirms or rejects each suggestion through the core's TransferService and can correct either decision later, while a read session only sees them.
logic_claims:
  - CLAIM-2C6B1349506EC441
semantic_clauses:
  - The built-in UI must list the suggested, confirmed and rejected transfers on a transfers page where an owner session confirms or rejects each suggestion through the core's TransferService and can correct either decision later, while a read session only sees them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cb63af88acdd5638b62f776a1247b305c1b22fb59d7ab615485750f8902228a9
semantic_inventory:
  - claim_key: CLAIM-2C6B1349506EC441
    claim_text: The built-in UI must list the suggested, confirmed and rejected transfers on a transfers page where an owner session confirms or rejects each suggestion through the core's TransferService and can correct either decision later, while a read session only sees them
    role: normative
    status: modeled
    span:
      start: 0
      end: 262
origin:
  kind: agent
  recorded_at: '2026-10-30T06:32:47.021Z'
id: REQ-ui-transfers-page
type: req
---
The built-in UI must list the suggested, confirmed and rejected transfers on a transfers page where an owner session confirms or rejects each suggestion through the core's TransferService and can correct either decision later, while a read session only sees them.

## Context

D-105: the owner confirms or rejects each suggestion in the UI. Refines the firm REQ-ui-transfer-confirmation with the page, the correction buttons and the read scope; the count on the performance card is the UI's reading.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
