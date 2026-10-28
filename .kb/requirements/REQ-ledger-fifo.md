---
title: Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.
status: open
text_ref: docs/DESIGN.md#31-system-of-record-and-views
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-31-system-of-record-and-views
  - lane:strict
semantic_text: Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.
logic_claims:
  - CLAIM-E9334EFBE9C77039
semantic_clauses:
  - Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5a375f7684f54d16e64ea40587821e4d3dddbd9651a4837876883c3eb83c9d53
semantic_inventory:
  - claim_key: CLAIM-E9334EFBE9C77039
    claim_text: Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date
    role: normative
    status: modeled
    span:
      start: 0
      end: 190
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:20.486Z'
id: REQ-ledger-fifo
type: req
---
Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.

## Context

DESIGN section 3.1 names FIFO tax lots as something OpenFief must get exactly right, and the glossary says lots are consumed FIFO. Tie-breaking by key makes the order deterministic for lots opened on the same day. Within one date, acquisitions are applied before disposals, because rows from the same day carry no reliable intra-day order. Lots are derived from the log (D-18), not stored. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Existing trackers can't hold what OpenFief has to get exactly right: FIFO tax lots priced in PLN

docs/DESIGN.md section 3.1

> A quantity acquired on a date at a cost; consumed FIFO.

docs/glossary.md (Lot)
