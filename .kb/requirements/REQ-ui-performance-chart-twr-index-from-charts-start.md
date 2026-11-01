---
title: The performance chart must show the time-weighted return as an index from 100 on the charts' start, computed on gross value.
status: open
text_ref: design-decisions:D-106#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-4
  - lane:strict
semantic_text: The performance chart must show the time-weighted return as an index from 100 on the charts' start, computed on gross value.
logic_claims:
  - CLAIM-373A0CAAA90AAA79
semantic_clauses:
  - The performance chart must show the time-weighted return as an index from 100 on the charts' start, computed on gross value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 847011c78a491e3063a24444beae3b94dcf9b84f9131e28f33283b275fd9cfb1
semantic_inventory:
  - claim_key: CLAIM-373A0CAAA90AAA79
    claim_text: The performance chart must show the time-weighted return as an index from 100 on the charts' start, computed on gross value
    role: normative
    status: modeled
    span:
      start: 0
      end: 123
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-106#4
  recorded_at: '2026-11-01T07:38:35.621Z'
id: REQ-ui-performance-chart-twr-index-from-charts-start
type: req
---
The performance chart must show the time-weighted return as an index from 100 on the charts' start, computed on gross value.

## Why this changed

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. REQ-ui-performance-chart-twr-index set the index to 100 at go-live, so this requirement replaces it.

## Context

D-61 and D-107 say how the time-weighted return is measured; D-106 separates the value and performance charts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 on the charts' start (D-138), computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#4
