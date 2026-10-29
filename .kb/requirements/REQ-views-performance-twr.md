---
title: The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places.
status: open
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
semantic_text: The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places.
logic_claims:
  - CLAIM-87C6F188421DACA1
semantic_clauses:
  - The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 332b0b45b66044f507483a0e7c27a891efc539fe534567fbcb7328852edb51e1
semantic_inventory:
  - claim_key: CLAIM-87C6F188421DACA1
    claim_text: The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places
    role: normative
    status: modeled
    span:
      start: 0
      end: 316
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:52.611Z'
id: REQ-views-performance-twr
type: req
---
The read API's performance series must chain each day's return of the portfolio's value with the day's net external cash flow taken at the end of the day, so that deposits and withdrawals are not returns, starting the index at 100 and rounding each day's factor and the running product half-even to 12 decimal places.

## Context

D-61 makes the portfolio's line a time-weighted return chained across deposits and withdrawals; D-53 starts both chart lines at 100. Valuations are end of day, so the end-of-day flow convention and the 12-place half-even precision are this context's choices, not the author's (GIPS/CFA daily-valued time-weighted return; Modified Dietz with the flow weighted zero).

## Source

> **[D-61]** Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows. Both lines are in the reporting currency (D-62). ...

docs/DESIGN.md#D-61
