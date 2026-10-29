---
title: The time-weighted return must count cash flows at the end of their day.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-1
  - lane:strict
text_ref: design-decisions:D-107#1
origin:
  kind: import
  ref: design-decisions:D-107#1
  recorded_at: '2026-10-29T10:16:05.080Z'
semantic_text: The time-weighted return must count cash flows at the end of their day.
logic_claims:
  - CLAIM-77F64097AEE1BDF7
semantic_clauses:
  - The time-weighted return must count cash flows at the end of their day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8496e952add1b02abf70d80e513db97383a8d61f85b89925ee67f75a48af2b2f
semantic_inventory:
  - claim_key: CLAIM-77F64097AEE1BDF7
    claim_text: The time-weighted return must count cash flows at the end of their day
    role: normative
    status: modeled
    span:
      start: 0
      end: 70
id: REQ-portfolio-twr-cash-flows-end-of-day
type: req
---
The time-weighted return must count cash flows at the end of their day.

## Context

D-61 measures the portfolio's line as a time-weighted return chained across deposits and withdrawals. D-107 records, as the author's decisions, conventions the implementation had already built for it and for the model portfolio; this is the first of them. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#1
