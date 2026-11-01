---
title: Every time chart, the portfolio's and each account's, must start on a date the owner can choose.
status: open
text_ref: design-decisions:D-138#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-138-1
  - lane:strict
semantic_text: Every time chart, the portfolio's and each account's, must start on a date the owner can choose.
logic_claims:
  - CLAIM-5BBB7D63750C88CD
semantic_clauses:
  - Every time chart, the portfolio's and each account's, must start on a date the owner can choose.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f99def3ff0ed4e6933e67d8d569128f48ceb6678583b4001adea422a22253e4d
semantic_inventory:
  - claim_key: CLAIM-5BBB7D63750C88CD
    claim_text: Every time chart, the portfolio's and each account's, must start on a date the owner can choose
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-138#1
  recorded_at: '2026-11-01T07:40:39.522Z'
id: REQ-ui-time-chart-start-owner-chosen
type: req
---
Every time chart, the portfolio's and each account's, must start on a date the owner can choose.

## Context

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. The time charts are D-30's value-over-time and performance charts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-138#1
