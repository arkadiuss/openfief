---
title: The performance chart must show the time-weighted return as an index from 100 at go-live, computed on gross value.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-4
  - lane:strict
text_ref: design-decisions:D-106#4
origin:
  kind: import
  ref: design-decisions:D-106#4
  recorded_at: '2026-10-29T10:14:26.979Z'
semantic_text: The performance chart must show the time-weighted return as an index from 100 at go-live, computed on gross value.
logic_claims:
  - CLAIM-DF55C06D50F06086
semantic_clauses:
  - The performance chart must show the time-weighted return as an index from 100 at go-live, computed on gross value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2d6152cd0545352cb164590b0b660093aecdf4cb832656d354fab7fcb5f4850f
semantic_inventory:
  - claim_key: CLAIM-DF55C06D50F06086
    claim_text: The performance chart must show the time-weighted return as an index from 100 at go-live, computed on gross value
    role: normative
    status: modeled
    span:
      start: 0
      end: 113
id: REQ-ui-performance-chart-twr-index
type: req
---
The performance chart must show the time-weighted return as an index from 100 at go-live, computed on gross value.

## Context

D-61 measures the portfolio's line as a time-weighted return. D-106 puts that line on its own performance chart, as an index from 100 at go-live computed on gross value, next to the model portfolio's line (REQ-views-model-line-on-performance-chart). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#4
