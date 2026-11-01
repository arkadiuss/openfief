---
title: Accounts with their whole history, or opened later, must not move the portfolio charts' default start.
status: open
text_ref: design-decisions:D-138#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-138-4
  - lane:strict
semantic_text: Accounts with their whole history, or opened later, must not move the portfolio charts' default start.
logic_claims:
  - CLAIM-EC7558717B9318EE
semantic_clauses:
  - Accounts with their whole history, or opened later, must not move the portfolio charts' default start.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d90520e09f53dc47529cc6d1c714ecd7000c64c1bba68aa7ec12007d943e7ba4
semantic_inventory:
  - claim_key: CLAIM-EC7558717B9318EE
    claim_text: Accounts with their whole history, or opened later, must not move the portfolio charts' default start
    role: normative
    status: modeled
    span:
      start: 0
      end: 101
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-138#4
  recorded_at: '2026-11-01T07:40:58.816Z'
id: REQ-ui-portfolio-chart-start-ignores-full-history
type: req
---
Accounts with their whole history, or opened later, must not move the portfolio charts' default start.

## Context

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. Only an account seeded from a starting balance (D-23) has history the record doesn't know; a bond account booked from its real purchases (D-98) or an XTB account started from its whole history (D-136) does not. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-138#4
