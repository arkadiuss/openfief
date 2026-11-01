---
title: The portfolio's charts must by default start on the latest day before which some account's history is unknown because it was seeded from a starting balance.
status: open
text_ref: design-decisions:D-138#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-138-3
  - lane:strict
semantic_text: The portfolio's charts must by default start on the latest day before which some account's history is unknown because it was seeded from a starting balance.
logic_claims:
  - CLAIM-F1B9D4C2688C3559
semantic_clauses:
  - The portfolio's charts must by default start on the latest day before which some account's history is unknown because it was seeded from a starting balance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 16859bcd3e37c2256f973396f42aa9eb22bde90ef9608ca1b2646a88174737c5
semantic_inventory:
  - claim_key: CLAIM-F1B9D4C2688C3559
    claim_text: The portfolio's charts must by default start on the latest day before which some account's history is unknown because it was seeded from a starting balance
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-138#3
  recorded_at: '2026-11-01T07:40:52.525Z'
id: REQ-ui-portfolio-chart-default-start
type: req
---
The portfolio's charts must by default start on the latest day before which some account's history is unknown because it was seeded from a starting balance.

## Context

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. D-23 seeds accounts from starting balances, so an account's history before the day it was seeded is unknown. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-138#3
