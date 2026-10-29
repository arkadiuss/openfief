---
title: The value-over-time chart must show money, cash included.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-2
  - lane:strict
text_ref: design-decisions:D-106#2
origin:
  kind: import
  ref: design-decisions:D-106#2
  recorded_at: '2026-10-29T10:13:00.448Z'
semantic_text: The value-over-time chart must show money, cash included.
logic_claims:
  - CLAIM-D1E9F64C03D2CEA4
semantic_clauses:
  - The value-over-time chart must show money, cash included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 56a861fbe60f304893c5ee94594838b815e4345deb7703df30878dd17a305e1a
semantic_inventory:
  - claim_key: CLAIM-D1E9F64C03D2CEA4
    claim_text: The value-over-time chart must show money, cash included
    role: normative
    status: modeled
    span:
      start: 0
      end: 56
id: REQ-ui-value-chart-money-with-cash
type: req
---
The value-over-time chart must show money, cash included.

## Context

D-106 separates the value chart from the performance chart (REQ-ui-value-performance-separate-charts). The value-over-time chart shows an amount of money, cash included, while performance is shown as an index on its own chart. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#2
