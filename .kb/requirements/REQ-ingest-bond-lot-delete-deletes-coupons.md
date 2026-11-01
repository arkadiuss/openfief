---
title: Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash.
status: open
text_ref: design-decisions:D-154#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-154-1
  - lane:strict
semantic_text: Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash.
logic_claims:
  - CLAIM-98D0E777122D264B
semantic_clauses:
  - Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c79c2658b2887fe8f34f73eae48f75332149de3346d65189c10717ef2d5c6b68
semantic_inventory:
  - claim_key: CLAIM-98D0E777122D264B
    claim_text: Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 104
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-154#1
  recorded_at: '2026-11-01T12:29:56.961Z'
id: REQ-ingest-bond-lot-delete-deletes-coupons
type: req
---
Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash.

## Context

The author recorded it on 2026-11-01. The decision gives the reason: those coupons were never paid. D-100 lets the owner delete a lot as a correction, which supersedes it and erases nothing (REQ-ingest-bond-lot-delete-supersedes). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Deleting a bond lot as a correction also deletes the coupons booked for it, and with them their cash, since those coupons were never paid.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-154#1
