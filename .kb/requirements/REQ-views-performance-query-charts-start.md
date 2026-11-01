---
title: The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running.
status: open
text_ref: docs/DESIGN.md#D-138
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-138
  - lane:strict
semantic_text: The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running.
logic_claims:
  - CLAIM-735604A2D566629B
semantic_clauses:
  - The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: de63e9b6166a8716fd1d6808ef6092784e5006099bd126138b34f910d80b3502
semantic_inventory:
  - claim_key: CLAIM-735604A2D566629B
    claim_text: The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running
    role: normative
    status: modeled
    span:
      start: 0
      end: 396
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:43.125Z'
id: REQ-views-performance-query-charts-start
type: req
---
The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running.

## Context

D-138 replaced go-live as the charts' first day with a start the owner chooses, whose default is the latest seeding day (D-23); D-36 makes pending revaluations visible. Counting an account never valued as valued through the day before its first transaction is kept from the requirement this one replaces. Recorded by the implementation agent of impl/chart-start for D-138; default firmness, not reviewed by the author yet.

## Why this changed

REQ-views-performance-query started the query at go-live, the first transaction's date, which D-138 no longer uses as the charts' start, so this requirement replaces it.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

docs/DESIGN.md#D-138
