---
title: The performance index must be 100 on a time chart's start date.
status: open
text_ref: design-decisions:D-138#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-138-2
  - lane:strict
semantic_text: The performance index must be 100 on a time chart's start date.
logic_claims:
  - CLAIM-6C88B546C7425416
semantic_clauses:
  - The performance index must be 100 on a time chart's start date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff0f4be1d8d199ad8f55a791c50f0d7ec951587f8323b5d3970098f5036bc71b
semantic_inventory:
  - claim_key: CLAIM-6C88B546C7425416
    claim_text: The performance index must be 100 on a time chart's start date
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-138#2
  recorded_at: '2026-11-01T07:40:45.907Z'
id: REQ-ui-performance-index-100-on-start
type: req
---
The performance index must be 100 on a time chart's start date.

## Context

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. D-106 draws the time-weighted return as an index, and D-53 starts the model's line at 100 on the same day. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-138#2
