---
title: The treasury-bonds connector must book each hand-entered lot on its purchase date as a deposit of its cost and a buy, in any order, keyed by series and purchase date, map each deleted lot to deletions of those keys and deliver each series' terms once; only in an account seeded with opening lots before must a lot on or before its go-live stay that opening lot.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: The treasury-bonds connector must book each hand-entered lot on its purchase date as a deposit of its cost and a buy, in any order, keyed by series and purchase date, map each deleted lot to deletions of those keys and deliver each series' terms once; only in an account seeded with opening lots before must a lot on or before its go-live stay that opening lot.
logic_claims:
  - CLAIM-EBBD43565155B328
semantic_clauses:
  - The treasury-bonds connector must book each hand-entered lot on its purchase date as a deposit of its cost and a buy, in any order, keyed by series and purchase date, map each deleted lot to deletions of those keys and deliver each series' terms once; only in an account seeded with opening lots before must a lot on or before its go-live stay that opening lot.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 11926561a282ad028f7d4e2f3c18681eea3a78f03f031075a002df995147218a
semantic_inventory:
  - claim_key: CLAIM-EBBD43565155B328
    claim_text: The treasury-bonds connector must book each hand-entered lot on its purchase date as a deposit of its cost and a buy, in any order, keyed by series and purchase date, map each deleted lot to deletions of those keys and deliver each series' terms once; only in an account seeded with opening lots before must a lot on or before its go-live stay that opening lot
    role: normative
    status: modeled
    span:
      start: 0
      end: 360
origin:
  kind: agent
  recorded_at: '2026-11-01T07:46:04.951Z'
id: REQ-treasurybonds-lot-on-purchase-date
type: req
---
The treasury-bonds connector must book each hand-entered lot on its purchase date as a deposit of its cost and a buy, in any order, keyed by series and purchase date, map each deleted lot to deletions of those keys and deliver each series' terms once; only in an account seeded with opening lots before must a lot on or before its go-live stay that opening lot.

## Context

D-98 was rewritten and D-119 removed: no go-live on the first entry. Keeping old opening lots on their keys is this implementation's reading (implementation question 54). Supersedes the go-live mapping. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

docs/DESIGN.md#D-98

## Why this changed

The author rewrote D-98 and removed D-119: a hand-entered account no longer goes live on its first entry, so every lot is booked on its real purchase date as a deposit and a buy, and only accounts seeded earlier keep their opening lots.
