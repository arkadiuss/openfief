---
title: A hand-entered treasury bond lot must be booked on its real purchase date, as a deposit of its cost and a buy.
status: open
text_ref: design-decisions:D-98#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-1
  - lane:strict
semantic_text: A hand-entered treasury bond lot must be booked on its real purchase date, as a deposit of its cost and a buy.
logic_claims:
  - CLAIM-827444029C7975F0
semantic_clauses:
  - A hand-entered treasury bond lot must be booked on its real purchase date, as a deposit of its cost and a buy.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eb7832651cabcfdfc660c2867c26445800fb346e31703d31b2c44ac7e3c9f5f6
semantic_inventory:
  - claim_key: CLAIM-827444029C7975F0
    claim_text: A hand-entered treasury bond lot must be booked on its real purchase date, as a deposit of its cost and a buy
    role: normative
    status: modeled
    span:
      start: 0
      end: 109
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-98#1
  recorded_at: '2026-11-01T07:36:35.861Z'
id: REQ-treasurybonds-lot-booked-on-purchase-date
type: req
---
A hand-entered treasury bond lot must be booked on its real purchase date, as a deposit of its cost and a buy.

## Why this changed

On 2026-11-01 the author rewrote D-98 and removed D-119: a hand-entered bond lot is booked on its real purchase date, because real purchase dates are the standard and keep the ledger a record of what happened. There are no opening lots at go-live for hand-entered bonds any more. REQ-treasurybonds-pre-go-live-opening-lot and REQ-treasurybonds-lot-on-go-live-day-opening (D-119) made a lot bought on or before the account's go-live an opening lot at go-live, and REQ-treasurybonds-post-go-live-deposit-buy booked only later lots as a deposit and a buy; every lot is now a deposit and a buy on its purchase date, so this requirement replaces all three.

## Context

D-9 has treasury bonds arrive through manual lot entry, and D-118 has the core hand each lot to the treasury-bonds connector with its account. D-100 sets what a lot costs. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#1
