---
title: Asked for one account, the performance query must compute that account's time-weighted return alone, from its own first day by default, counting each of its deposits and withdrawals as a cash flow, a transfer to or from another of the owner's accounts included.
status: open
text_ref: docs/DESIGN.md#D-138
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-138
  - lane:strict
semantic_text: Asked for one account, the performance query must compute that account's time-weighted return alone, from its own first day by default, counting each of its deposits and withdrawals as a cash flow, a transfer to or from another of the owner's accounts included.
logic_claims:
  - CLAIM-6E1DA03A0FA1BA2A
semantic_clauses:
  - Asked for one account, the performance query must compute that account's time-weighted return alone, from its own first day by default, counting each of its deposits and withdrawals as a cash flow, a transfer to or from another of the owner's accounts included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb6ff0809ec8108d5a2f2f4f30ba0f4dcb77f3684998ff8ed3fba3bd78e3c151
semantic_inventory:
  - claim_key: CLAIM-6E1DA03A0FA1BA2A
    claim_text: Asked for one account, the performance query must compute that account's time-weighted return alone, from its own first day by default, counting each of its deposits and withdrawals as a cash flow, a transfer to or from another of the owner's accounts included
    role: normative
    status: modeled
    span:
      start: 0
      end: 260
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:53.567Z'
id: REQ-views-account-performance
type: req
---
Asked for one account, the performance query must compute that account's time-weighted return alone, from its own first day by default, counting each of its deposits and withdrawals as a cash flow, a transfer to or from another of the owner's accounts included.

## Context

D-138 gives each account its own time charts, starting by default on the account's first day. Counting a transfer between the owner's accounts as a cash flow of the one account is this implementation's reading: D-61 leaves transfers out of the portfolio's flows because the money stays in the portfolio, which is not true of one account. Recorded by the implementation agent of impl/chart-start for D-138; default firmness, not reviewed by the author yet.

## Source

> Every time chart, the portfolio's and each account's, starts on a date the owner can choose, and the performance index is 100 on it. The portfolio's default is the latest day before which some account's history is unknown because it was seeded from a starting balance (D-23); accounts with their whole history, or opened later, don't move it, and the chart marks where tracking began. An account's default is its own first day.

docs/DESIGN.md#D-138
