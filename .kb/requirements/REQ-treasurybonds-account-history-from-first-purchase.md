---
title: A hand-entered bond account's history must start at its first purchase.
status: open
text_ref: design-decisions:D-98#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-5
  - lane:strict
semantic_text: A hand-entered bond account's history must start at its first purchase.
logic_claims:
  - CLAIM-BC648DC017B2297B
semantic_clauses:
  - A hand-entered bond account's history must start at its first purchase.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2d5465c72070b1b4339b45fc7cb2c2efd2bfbfdd47a1bbd542c7f116e111c3d1
semantic_inventory:
  - claim_key: CLAIM-BC648DC017B2297B
    claim_text: A hand-entered bond account's history must start at its first purchase
    role: normative
    status: modeled
    span:
      start: 0
      end: 70
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-98#5
  recorded_at: '2026-11-01T07:38:02.935Z'
id: REQ-treasurybonds-account-history-from-first-purchase
type: req
---
A hand-entered bond account's history must start at its first purchase.

## Why this changed

On 2026-11-01 the author rewrote D-98 and removed D-119: a hand-entered bond lot is booked on its real purchase date, because real purchase dates are the standard and keep the ledger a record of what happened. There are no opening lots at go-live for hand-entered bonds any more. REQ-ingest-bond-account-live-on-first-lot (D-119) had a hand-entered bond account go live on the day its first lot was entered; with lots booked on their real purchase dates, its history starts at its first purchase instead, so this requirement replaces it.

## Context

D-23's go-live, the day an account was seeded from opening lots, no longer applies to hand-entered bond accounts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#5
