---
title: Hand-entered treasury bond lots must be accepted one at a time and in any order.
status: open
text_ref: design-decisions:D-98#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-4
  - lane:strict
semantic_text: Hand-entered treasury bond lots must be accepted one at a time and in any order.
logic_claims:
  - CLAIM-6B4B2B07E46238EE
semantic_clauses:
  - Hand-entered treasury bond lots must be accepted one at a time and in any order.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1e183db7456ccd78bf8e430fb994aec5e8c6e47026a02c0546879c5c14016bed
semantic_inventory:
  - claim_key: CLAIM-6B4B2B07E46238EE
    claim_text: Hand-entered treasury bond lots must be accepted one at a time and in any order
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-98#4
  recorded_at: '2026-11-01T07:37:53.075Z'
id: REQ-treasurybonds-lots-any-order
type: req
---
Hand-entered treasury bond lots must be accepted one at a time and in any order.

## Context

On 2026-11-01 the author rewrote D-98 and removed D-119: a hand-entered bond lot is booked on its real purchase date, because real purchase dates are the standard and keep the ledger a record of what happened. There are no opening lots at go-live for hand-entered bonds any more. Because each lot is booked on its own purchase date and is idempotent per series and purchase date (REQ-treasurybonds-lot-idempotent-series-date), the owner can enter an older lot after a newer one. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#4
