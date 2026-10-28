---
title: An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected.
logic_claims:
  - CLAIM-DB616724C3841645
semantic_clauses:
  - An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 11ac6fb38139d6a8882c04085c127024c2a9c3dd297ebc1b7764b9b2d077de87
semantic_inventory:
  - claim_key: CLAIM-DB616724C3841645
    claim_text: An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-28T05:25:00.343Z'
id: REQ-ledger-opening-lot-entry
type: req
---
An opening lot must be recorded on the go-live date with the date the lot was first acquired and its cost basis in the original currency, and an open date after go-live must be rejected.

## Context

Refines the firm requirement REQ-AUTO-3C2A493C661181C7 (D-23): accounts are seeded from each broker's current lot-level positions, with transactions from go-live. The lot's original open date and cost are what FIFO order and later tax reports need, while the transaction itself belongs to go-live so the log still starts there. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md D-23
