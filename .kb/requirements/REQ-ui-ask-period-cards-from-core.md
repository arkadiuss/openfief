---
title: The Ask box's performance card for a month or two days must show the time-weighted return, the portfolio's value at the start and the end and the deposits and withdrawals, and its value card for a period the gross, net and liquidation value at the start and the end with the change in money and in percent and the net deposits, every figure taken from the read API's GetPeriodChange and none computed by the box.
status: open
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: the Ask box performance and value tools and commands answer for a bounded period, a gap against D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T13:36:17.484Z'
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: The Ask box's performance card for a month or two days must show the time-weighted return, the portfolio's value at the start and the end and the deposits and withdrawals, and its value card for a period the gross, net and liquidation value at the start and the end with the change in money and in percent and the net deposits, every figure taken from the read API's GetPeriodChange and none computed by the box.
tags:
  - ui
  - ask
  - period
  - d-194
  - d-203
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ace37302002d765867e553a6ab6446b7ec644152b99660c7ba1344ecc0c79c0b
semantic_inventory:
  - claim_key: CLAIM-67E763A1E60DA30C
    claim_text: The Ask box's performance card for a month or two days must show the time-weighted return, the portfolio's value at the start and the end and the deposits and withdrawals, and its value card for a period the gross, net and liquidation value at the start and the end with the change in money and in percent and the net deposits, every figure taken from the read API's GetPeriodChange and none computed by the box
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 411
logic_claims:
  - CLAIM-67E763A1E60DA30C
id: REQ-ui-ask-period-cards-from-core
type: req
---
The Ask box's performance card for a month or two days must show the time-weighted return, the portfolio's value at the start and the end and the deposits and withdrawals, and its value card for a period the gross, net and liquidation value at the start and the end with the change in money and in percent and the net deposits, every figure taken from the read API's GetPeriodChange and none computed by the box.

## Context

D-194 lists performance over a period and the value on a date as commands of the Ask box, and D-203 lets its model call only the same read-only answerers. Before this, only periods that start on a fixed day (this year, since a date, a calendar year) could be asked, so a question about one month, such as why the value or the return changed in November 2025, could not be answered. The read API answers a bounded period first (REQ-views-period-change), and the box shows its figures. Recorded by the implementation agent for the design session, which approved the gap on the author's behalf; default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203