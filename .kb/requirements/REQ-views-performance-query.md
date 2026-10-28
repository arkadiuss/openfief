---
title: The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running.
logic_claims:
  - CLAIM-8CE6227208D074E7
semantic_clauses:
  - The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b68ef7bc9da6c8347c9507ecea69350410f5d731de3564cba124fec6c4c62f7b
semantic_inventory:
  - claim_key: CLAIM-8CE6227208D074E7
    claim_text: The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running
    role: normative
    status: modeled
    span:
      start: 0
      end: 325
origin:
  kind: agent
  recorded_at: '2026-10-28T20:55:20.374Z'
id: REQ-views-performance-query
type: req
---
The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running.

## Context

D-30's value-over-time chart starts at go-live; D-36 makes pending revaluations visible. The defaults let the UI ask without dates.

## Source

> **[D-30]** The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md#D-30
