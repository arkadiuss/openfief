---
title: A cash balance the owner enters must be stored as a balance assertion.
status: open
text_ref: design-decisions:D-158#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-2
  - lane:strict
semantic_text: A cash balance the owner enters must be stored as a balance assertion.
logic_claims:
  - CLAIM-EC1DFCA97D36B714
semantic_clauses:
  - A cash balance the owner enters must be stored as a balance assertion.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 40c727ba0af5885f989c04b79654d836580addb2b87b9621bed13d05fedc7dbe
semantic_inventory:
  - claim_key: CLAIM-EC1DFCA97D36B714
    claim_text: A cash balance the owner enters must be stored as a balance assertion
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#2
  recorded_at: '2026-11-01T12:31:25.977Z'
id: REQ-ledger-owner-cash-stored-as-assertion
type: req
---
A cash balance the owner enters must be stored as a balance assertion.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. D-18 stores broker position snapshots as balance assertions (REQ-AUTO-D9FBE33256704D04); the owner's entry is checked the same way. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#2
