---
title: Each hand-entered treasury bond lot must be idempotent per series and purchase date.
status: open
text_ref: design-decisions:D-98#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-2
  - lane:strict
semantic_text: Each hand-entered treasury bond lot must be idempotent per series and purchase date.
logic_claims:
  - CLAIM-67D75FDE55F0E273
semantic_clauses:
  - Each hand-entered treasury bond lot must be idempotent per series and purchase date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c946ebc8e1d2b802e6a1a32ca1d52056838033e7505b9bda6a424684b0fbf3a
semantic_inventory:
  - claim_key: CLAIM-67D75FDE55F0E273
    claim_text: Each hand-entered treasury bond lot must be idempotent per series and purchase date
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-98#2
  recorded_at: '2026-11-01T07:37:45.526Z'
id: REQ-treasurybonds-lot-idempotent-series-date
type: req
---
Each hand-entered treasury bond lot must be idempotent per series and purchase date.

## Why this changed

On 2026-11-01 the author rewrote D-98 and removed D-119: a hand-entered bond lot is booked on its real purchase date, because real purchase dates are the standard and keep the ledger a record of what happened. There are no opening lots at go-live for hand-entered bonds any more. REQ-ledger-manual-opening-lots-one-at-a-time let a manually entered account receive opening lots one at a time, each idempotent per series and purchase date; lots are no longer opening lots, so this requirement keeps the idempotency for the lots themselves and replaces it.

## Context

Entering the same lot again, or a connector delivering it again on a later sync, must not book it twice. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#2
