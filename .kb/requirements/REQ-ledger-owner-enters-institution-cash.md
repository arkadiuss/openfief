---
title: The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's.
status: open
text_ref: design-decisions:D-158#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-158-1
  - lane:strict
semantic_text: The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's.
logic_claims:
  - CLAIM-5706E766B6474CE1
semantic_clauses:
  - The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 04579b6892f3633c1a1874ccd5090fbdfab13766c23a22b8d138c6892d72bf24
semantic_inventory:
  - claim_key: CLAIM-5706E766B6474CE1
    claim_text: The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's
    role: normative
    status: modeled
    span:
      start: 0
      end: 135
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-158#1
  recorded_at: '2026-11-01T12:31:16.176Z'
id: REQ-ledger-owner-enters-institution-cash
type: req
---
The owner must be able to enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's.

## Context

The author recorded it on 2026-11-01 as the cash checkpoint. Reason not stated. The decision names the bond registry's and a bank's balances as examples; hand-entered bond accounts (D-98, D-135) are one such case. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-158#1
