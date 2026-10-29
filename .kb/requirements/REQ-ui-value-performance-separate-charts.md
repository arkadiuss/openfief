---
title: Value and performance must be separate charts.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-1
  - lane:strict
text_ref: design-decisions:D-106#1
origin:
  kind: import
  ref: design-decisions:D-106#1
  recorded_at: '2026-10-29T10:12:20.755Z'
semantic_text: Value and performance must be separate charts.
logic_claims:
  - CLAIM-0210C46C96730463
semantic_clauses:
  - Value and performance must be separate charts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 269127ffabe882f65cc043db04b4e6b66f692f20f5d74820c606a539ea9e246e
semantic_inventory:
  - claim_key: CLAIM-0210C46C96730463
    claim_text: Value and performance must be separate charts
    role: normative
    status: modeled
    span:
      start: 0
      end: 45
id: REQ-ui-value-performance-separate-charts
type: req
---
Value and performance must be separate charts.

## Context

Before D-106, the value-over-time chart of D-30 also carried the model portfolio's line from D-53. D-106 splits value and performance into separate charts, and D-30 now names three charts (REQ-ui-performance-chart). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#1
