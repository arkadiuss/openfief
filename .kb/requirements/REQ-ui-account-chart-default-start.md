---
title: An account's time charts must by default start on the account's own first day.
status: open
text_ref: design-decisions:D-138#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-138-6
  - lane:strict
semantic_text: An account's time charts must by default start on the account's own first day.
logic_claims:
  - CLAIM-72BEE22D890A28DE
semantic_clauses:
  - An account's time charts must by default start on the account's own first day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5c22a3b2806ad57bcd2b51ca0a6465f978dab180c03904346494b977733a00a5
semantic_inventory:
  - claim_key: CLAIM-72BEE22D890A28DE
    claim_text: An account's time charts must by default start on the account's own first day
    role: normative
    status: modeled
    span:
      start: 0
      end: 77
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-138#6
  recorded_at: '2026-11-01T07:41:10.963Z'
id: REQ-ui-account-chart-default-start
type: req
---
An account's time charts must by default start on the account's own first day.

## Context

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. An account's default start doesn't depend on the other accounts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-138#6
