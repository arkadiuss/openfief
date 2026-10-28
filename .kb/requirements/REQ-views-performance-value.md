---
title: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
status: open
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
semantic_text: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
logic_claims:
  - CLAIM-F83DC3437EE641BE
semantic_clauses:
  - Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c46e5da11457a5182ef228272587eba4417e34e79b84dab710dbb91463ab260a
semantic_inventory:
  - claim_key: CLAIM-F83DC3437EE641BE
    claim_text: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation
    role: normative
    status: modeled
    span:
      start: 0
      end: 334
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:59.815Z'
id: REQ-views-performance-value
type: req
---
Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.

## Context

D-61 computes daily returns from valuations; stored valuations hold no cash, which the read API's totals left out, so the series adds it the way the allocation does (valuation.ValueRounding, half up). Opening lots of an account seeded after go-live are new money, valued at market, not at cost.

## Source

> **[D-61]** Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows. Both lines are in the reporting currency (D-62). ...

docs/DESIGN.md#D-61
