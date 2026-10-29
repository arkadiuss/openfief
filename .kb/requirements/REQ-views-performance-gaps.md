---
title: The performance series must stop before the first day that is after an account's valued-through date, on which a holding lacks one complete valuation of its current quantity, or on which cash, a cash flow or money in transit has no FX rate, and must name which gap stopped it.
status: open
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
semantic_text: The performance series must stop before the first day that is after an account's valued-through date, on which a holding lacks one complete valuation of its current quantity, or on which cash, a cash flow or money in transit has no FX rate, and must name which gap stopped it.
logic_claims:
  - CLAIM-B09018DB4B28917A
semantic_clauses:
  - The performance series must stop before the first day that is after an account's valued-through date, on which a holding lacks one complete valuation of its current quantity, or on which cash, a cash flow or money in transit has no FX rate, and must name which gap stopped it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3cdae27b2074fbc549e0d295dcd7e2e8e9284cffbde66c428a2f54978bb9915c
semantic_inventory:
  - claim_key: CLAIM-B09018DB4B28917A
    claim_text: The performance series must stop before the first day that is after an account's valued-through date, on which a holding lacks one complete valuation of its current quantity, or on which cash, a cash flow or money in transit has no FX rate, and must name which gap stopped it
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-10-28T20:55:13.910Z'
id: REQ-views-performance-gaps
type: req
---
The performance series must stop before the first day that is after an account's valued-through date, on which a holding lacks one complete valuation of its current quantity, or on which cash, a cash flow or money in transit has no FX rate, and must name which gap stopped it.

## Context

D-36 reports per account the date it is valued through; a day missing an input would draw a false drop in the chart, so the series ends explicitly at the last fully valued day. Asked for by the orchestrator.

## Source

> **[D-61]** Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows. Both lines are in the reporting currency (D-62). ...

docs/DESIGN.md#D-61
