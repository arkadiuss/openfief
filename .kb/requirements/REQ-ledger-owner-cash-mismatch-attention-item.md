---
title: A mismatch between a cash balance the owner entered and the derived cash must raise an attention item.
status: open
text_ref: design-decisions:D-158#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-3
  - lane:strict
semantic_text: A mismatch between a cash balance the owner entered and the derived cash must raise an attention item.
logic_claims:
  - CLAIM-6EE46CE5F1D06296
semantic_clauses:
  - A mismatch between a cash balance the owner entered and the derived cash must raise an attention item.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d014713b95bc3f14a7f09cf6439ca9c53c1cab1b5031591a4ab34f741eefe080
semantic_inventory:
  - claim_key: CLAIM-6EE46CE5F1D06296
    claim_text: A mismatch between a cash balance the owner entered and the derived cash must raise an attention item
    role: normative
    status: modeled
    span:
      start: 0
      end: 101
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#3
  recorded_at: '2026-11-01T12:31:34.827Z'
id: REQ-ledger-owner-cash-mismatch-attention-item
type: req
---
A mismatch between a cash balance the owner entered and the derived cash must raise an attention item.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. Flagged balance assertions raise attention items in D-130's store (REQ-core-attention-item-checks). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#3
