---
title: The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.
status: open
text_ref: docs/DESIGN.md#D-158
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-158
  - lane:strict
semantic_text: The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.
logic_claims:
  - CLAIM-72EE8F3A3344A2BE
semantic_clauses:
  - The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 467c33e8c673acb5929eb4ec0be3eaa639d5b7ae839b882510dcad20380c2120
semantic_inventory:
  - claim_key: CLAIM-72EE8F3A3344A2BE
    claim_text: The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then
    role: normative
    status: modeled
    span:
      start: 0
      end: 246
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:17.754Z'
id: REQ-ledger-cash-checkpoint-service
type: req
---
The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.

## Context

D-158 lets the owner enter cash and book differences; D-79 puts each kind of owner entry behind one owner-only service. The refusals and their codes are this implementation's reading, following the TransferService's conventions; D-163 makes every owner service ask the authorizer whether the owner may change the account. Asked by the author through D-158, D-79 and D-163.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

docs/DESIGN.md#D-158
