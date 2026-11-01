---
title: 'From a cash balance the owner entered, the owner must be able to book the difference as an ordinary manual transaction: a deposit or withdrawal, or cash income when it is income or a fee.'
status: open
text_ref: design-decisions:D-158#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-4
  - lane:strict
semantic_text: 'From a cash balance the owner entered, the owner must be able to book the difference as an ordinary manual transaction: a deposit or withdrawal, or cash income when it is income or a fee.'
logic_claims:
  - CLAIM-90F960B6B0E62628
semantic_clauses:
  - 'From a cash balance the owner entered, the owner must be able to book the difference as an ordinary manual transaction: a deposit or withdrawal, or cash income when it is income or a fee.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4f88d47e1a802b363c9c8494224962ad8ce66a3d2482ea7640cfbc08aa691a3b
semantic_inventory:
  - claim_key: CLAIM-90F960B6B0E62628
    claim_text: 'From a cash balance the owner entered, the owner must be able to book the difference as an ordinary manual transaction: a deposit or withdrawal, or cash income when it is income or a fee'
    role: normative
    status: modeled
    span:
      start: 0
      end: 186
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#4
  recorded_at: '2026-11-01T12:31:43.111Z'
id: REQ-ledger-owner-cash-difference-booked
type: req
---
From a cash balance the owner entered, the owner must be able to book the difference as an ordinary manual transaction: a deposit or withdrawal, or cash income when it is income or a fee.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. Cash income is D-152's kind for income no holding produces (REQ-ledger-cash-income-kind). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#4
