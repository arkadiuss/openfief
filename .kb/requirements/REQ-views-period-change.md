---
title: GetPeriodChange must return, for a period of days, the portfolio's gross, net and liquidation values at the start and the end with their change in money and in percent, the period's time-weighted return and the deposits and withdrawals in it, all computed by the core, and must refuse a period the portfolio is not valued for.
status: open
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: the Ask box''s performance and value tools answer for a bounded period, a gap against D-194 (D-203: the read API answers first), 2026-11-15'
  recorded_at: '2026-11-15T12:27:21.272Z'
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: GetPeriodChange must return, for a period of days, the portfolio's gross, net and liquidation values at the start and the end with their change in money and in percent, the period's time-weighted return and the deposits and withdrawals in it, all computed by the core, and must refuse a period the portfolio is not valued for.
tags:
  - views
  - period
  - d-194
  - d-203
  - d-61
  - d-77
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bd964127a8349906ffe3289d5f5bd5ea474beb3f4f9cd86516d5cd5351be43dd
semantic_inventory:
  - claim_key: CLAIM-21BB6E0DCE5248F9
    claim_text: GetPeriodChange must return, for a period of days, the portfolio's gross, net and liquidation values at the start and the end with their change in money and in percent, the period's time-weighted return and the deposits and withdrawals in it, all computed by the core, and must refuse a period the portfolio is not valued for
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 325
logic_claims:
  - CLAIM-21BB6E0DCE5248F9
id: REQ-views-period-change
type: req
---
GetPeriodChange must return, for a period of days, the portfolio's gross, net and liquidation values at the start and the end with their change in money and in percent, the period's time-weighted return and the deposits and withdrawals in it, all computed by the core, and must refuse a period the portfolio is not valued for.

## Context

The Ask box's performance and value tools could answer a date or a period that starts on a fixed day, but not a bounded one, such as November 2025, which D-194's performance over a period and value on a date name. D-203 lets the box's tools be only the read API's answerers, so the read API answers a period first, additively: one method, so a client shows the figures without arithmetic of its own. That the start is the end of the day before the first day, and that a day's deposits and withdrawals are netted before they are counted, are the implementation's readings. Recorded by the implementation agent for the design session, which approved the gap on the author's behalf; default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203, D-61 (the time-weighted return), D-77 (the three values)