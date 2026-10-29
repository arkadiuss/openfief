---
title: Valuation totals must include cash.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-5
  - lane:strict
text_ref: design-decisions:D-106#5
origin:
  kind: import
  ref: design-decisions:D-106#5
  recorded_at: '2026-10-29T10:15:07.420Z'
semantic_text: Valuation totals must include cash.
logic_claims:
  - CLAIM-A5F11DC63CA24145
semantic_clauses:
  - Valuation totals must include cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a5ed5859b21081bd10b36186ae708f672a291adf161d49a995a6aa513d58f8c5
semantic_inventory:
  - claim_key: CLAIM-A5F11DC63CA24145
    claim_text: Valuation totals must include cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 34
id: REQ-valuation-totals-include-cash
type: req
---
Valuation totals must include cash.

## Context

D-106 has the value-over-time chart show money with cash included (REQ-ui-value-chart-money-with-cash) and states that valuation totals include cash. Reason not stated beyond that. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#5
