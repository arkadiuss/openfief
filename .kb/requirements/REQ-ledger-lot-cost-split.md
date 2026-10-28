---
title: Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost.
logic_claims:
  - CLAIM-1F085A54BEC551B4
semantic_clauses:
  - Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a966458c351bb191e207091144c27470b5f333f5005f545c37b622c2412493a2
semantic_inventory:
  - claim_key: CLAIM-1F085A54BEC551B4
    claim_text: Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:25.009Z'
id: REQ-ledger-lot-cost-split
type: req
---
Consuming part of a lot must give the consumed units the lot's cost times their share of its units, rounded half-even to 2 decimal places, and leave the exact rest in the lot, so the consumed and remaining cost always add up to the original cost.

## Context

Partial sales are common, and an exact split of a cost in decimal is not always possible (100.00 over 3 units). Rounding only the consumed part and keeping the exact remainder in the lot means money is never created or lost, and the last unit sold takes whatever is left. Two places and half-even rounding are Claude Code's choice for the currencies in scope (PLN, USD, EUR), not stated by the author; costs stay in their original currency (D-20). Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Every amount keeps its original currency.

docs/DESIGN.md D-20

> Domain logic is pure functions with table tests. No `float64` for money.

AGENTS.md#hard-rules
