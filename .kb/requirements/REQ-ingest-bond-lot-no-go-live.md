---
title: The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date.
logic_claims:
  - CLAIM-BBF2D4BFEAC92A88
semantic_clauses:
  - The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af8fe958599c5a04055cb23f6a5b3efee23b7bf32aa62b6093cf7ee9fde54f57
semantic_inventory:
  - claim_key: CLAIM-BBF2D4BFEAC92A88
    claim_text: The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
origin:
  kind: agent
  recorded_at: '2026-11-01T07:46:11.567Z'
id: REQ-ingest-bond-lot-no-go-live
type: req
---
The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date.

## Context

D-98 rewritten, D-119 removed: an account no longer goes live on its first entry. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> A hand-entered treasury bond lot is booked on its real purchase date, as a deposit of its cost and a buy, so the account's history starts at its first purchase and every day since is valued from the series' terms and the published readings. Each lot is idempotent per series and purchase date, so lots can be entered one at a time and in any order.

docs/DESIGN.md#D-98

## Why this changed

The author rewrote D-98 and removed D-119: the core no longer sends today as the go-live of an account not seeded, because its lots are booked on their purchase dates and it never goes live.
