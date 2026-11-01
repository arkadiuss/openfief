---
title: The difference from a cash balance the owner entered must be booked as a deposit or withdrawal by default.
status: open
text_ref: design-decisions:D-158#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-5
  - lane:strict
semantic_text: The difference from a cash balance the owner entered must be booked as a deposit or withdrawal by default.
logic_claims:
  - CLAIM-6E090352F9F95EC0
semantic_clauses:
  - The difference from a cash balance the owner entered must be booked as a deposit or withdrawal by default.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5ad5b2ba0ebb1d8c0f758c585412032ca1c8499c85dcb94a085237706f685480
semantic_inventory:
  - claim_key: CLAIM-6E090352F9F95EC0
    claim_text: The difference from a cash balance the owner entered must be booked as a deposit or withdrawal by default
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#5
  recorded_at: '2026-11-01T12:31:51.712Z'
id: REQ-ledger-owner-cash-difference-default-deposit
type: req
---
The difference from a cash balance the owner entered must be booked as a deposit or withdrawal by default.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. The decision gives the reason: an unexplained difference never shows as return, since performance leaves deposits and withdrawals out (D-61). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#5
