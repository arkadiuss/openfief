---
title: The Ask box's performance and value commands must understand a calendar month, a calendar year (value only) or two days as the period they ask about, with a month written in English or as YYYY-MM and a day as an ISO or day.month.year date, and must refuse a month or a day that does not exist.
status: open
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: the Ask box performance and value tools and commands answer for a bounded period, a gap against D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T13:36:14.758Z'
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: The Ask box's performance and value commands must understand a calendar month, a calendar year (value only) or two days as the period they ask about, with a month written in English or as YYYY-MM and a day as an ISO or day.month.year date, and must refuse a month or a day that does not exist.
tags:
  - ui
  - ask
  - period
  - d-194
  - d-203
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6b36d224f92e1b059cb17508d79e51fb2c753b3606d553e51836f82e4e392576
semantic_inventory:
  - claim_key: CLAIM-7BFDB9E92FB7BA31
    claim_text: The Ask box's performance and value commands must understand a calendar month, a calendar year (value only) or two days as the period they ask about, with a month written in English or as YYYY-MM and a day as an ISO or day.month.year date, and must refuse a month or a day that does not exist
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 292
    subjects:
      - ask_box.period_commands
logic_claims:
  - CLAIM-7BFDB9E92FB7BA31
id: REQ-ui-ask-period-commands
type: req
---
The Ask box's performance and value commands must understand a calendar month, a calendar year (value only) or two days as the period they ask about, with a month written in English or as YYYY-MM and a day as an ISO or day.month.year date, and must refuse a month or a day that does not exist.

## Context

D-194 lists performance over a period and the value on a date as commands of the Ask box, and D-203 lets its model call only the same read-only answerers. Before this, only periods that start on a fixed day (this year, since a date, a calendar year) could be asked, so a question about one month, such as why the value or the return changed in November 2025, could not be answered. The read API answers a bounded period first (REQ-views-period-change), and the box shows its figures. Recorded by the implementation agent for the design session, which approved the gap on the author's behalf; default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203