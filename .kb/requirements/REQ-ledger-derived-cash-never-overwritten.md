---
title: Nothing may overwrite the derived cash balance.
status: open
text_ref: design-decisions:D-158#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-6
  - lane:strict
semantic_text: Nothing may overwrite the derived cash balance.
logic_claims:
  - CLAIM-1F64BFB7D1DE43E7
semantic_clauses:
  - Nothing may overwrite the derived cash balance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bd5e3e7b92dd57ceb263fa3106ea2007e26630b6e31f071e0304492a073f3bac
semantic_inventory:
  - claim_key: CLAIM-1F64BFB7D1DE43E7
    claim_text: Nothing may overwrite the derived cash balance
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 46
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#6
  recorded_at: '2026-11-01T12:31:59.838Z'
id: REQ-ledger-derived-cash-never-overwritten
type: req
---
Nothing may overwrite the derived cash balance.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. D-18 derives holdings and cash from the append-only transaction log; a difference is booked as a transaction instead. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#6
