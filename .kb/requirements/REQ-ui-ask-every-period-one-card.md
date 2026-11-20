---
title: Every performance period the Ask box answers, in a typed command and in the model's performance tool, this year, this month, the last 12 months, last year, a calendar year, a first day onward, everything since tracking began, a calendar month and two days, must be resolved to its first and last day and answered with the period card from the read API's GetPeriodChange, so that each period's return runs from the end of the day before its first day to the end of its last day and its card shows the start and end values and the deposits and withdrawals.
status: open
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: Every performance period the Ask box answers, in a typed command and in the model's performance tool, this year, this month, the last 12 months, last year, a calendar year, a first day onward, everything since tracking began, a calendar month and two days, must be resolved to its first and last day and answered with the period card from the read API's GetPeriodChange, so that each period's return runs from the end of the day before its first day to the end of its last day and its card shows the start and end values and the deposits and withdrawals.
tags:
  - ui
  - ask
  - period
  - d-61
  - d-194
  - d-203
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: every performance period answers with the period card and one convention, a gap against D-61, D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T15:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f4ea5d7cc7a5fccdc23f5e8de67f9b8bc09fdbf8fb2b55f14704ce781c50759c
semantic_inventory:
  - claim_key: CLAIM-0595AD7776DFFC2D
    claim_text: Every performance period the Ask box answers, in a typed command and in the model's performance tool, this year, this month, the last 12 months, last year, a calendar year, a first day onward, everything since tracking began, a calendar month and two days, must be resolved to its first and last day and answered with the period card from the read API's GetPeriodChange, so that each period's return runs from the end of the day before its first day to the end of its last day and its card shows the start and end values and the deposits and withdrawals
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 553
    subjects:
      - ask_box.period_cards
logic_claims:
  - CLAIM-0595AD7776DFFC2D
id: REQ-ui-ask-every-period-one-card
type: req
---
Every performance period the Ask box answers, in a typed command and in the model's performance tool, this year, this month, the last 12 months, last year, a calendar year, a first day onward, everything since tracking began, a calendar month and two days, must be resolved to its first and last day and answered with the period card from the read API's GetPeriodChange, so that each period's return runs from the end of the day before its first day to the end of its last day and its card shows the start and end values and the deposits and withdrawals.

## Context

D-61 chains the time-weighted return across cash flows, D-194 lists performance over a period as a command of the Ask box and D-203 lets its model call only the same read-only answerers. Before this, only a month and two days used GetPeriodChange: the other periods answered from the performance series, whose index put 100 on the first day's own close, so the first day did not count and the card lacked the start and end values and the deposits. A model then made a second call for the values (bench case A9), and "performance in 2025" and "in November 2025" started on different conventions. The design session approved, on the author's behalf, one convention and the one period card for every performance period. Default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203, D-61