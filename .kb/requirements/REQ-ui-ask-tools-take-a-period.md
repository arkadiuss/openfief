---
title: The performance tool must accept a calendar month or a first and a last day and the value tool a period of a first and a last day, each mapped to the query a typed command of that period makes, and must refuse a period without both days and a value asked on a date and over a period at once.
status: open
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: the Ask box performance and value tools and commands answer for a bounded period, a gap against D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T13:36:22.844Z'
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: The performance tool must accept a calendar month or a first and a last day and the value tool a period of a first and a last day, each mapped to the query a typed command of that period makes, and must refuse a period without both days and a value asked on a date and over a period at once.
tags:
  - ui
  - ask
  - period
  - d-194
  - d-203
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2bdea7ec604b353314e7b2a2a073bac3e80e4ffc8a92ad62164f55bc41bb8c8a
semantic_inventory:
  - claim_key: CLAIM-3FB3250217DF6AEA
    claim_text: The performance tool must accept a calendar month or a first and a last day and the value tool a period of a first and a last day, each mapped to the query a typed command of that period makes, and must refuse a period without both days and a value asked on a date and over a period at once
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 290
logic_claims:
  - CLAIM-3FB3250217DF6AEA
id: REQ-ui-ask-tools-take-a-period
type: req
---
The performance tool must accept a calendar month or a first and a last day and the value tool a period of a first and a last day, each mapped to the query a typed command of that period makes, and must refuse a period without both days and a value asked on a date and over a period at once.

## Context

D-194 lists performance over a period and the value on a date as commands of the Ask box, and D-203 lets its model call only the same read-only answerers. Before this, only periods that start on a fixed day (this year, since a date, a calendar year) could be asked, so a question about one month, such as why the value or the return changed in November 2025, could not be answered. The read API answers a bounded period first (REQ-views-period-change), and the box shows its figures. Recorded by the implementation agent for the design session, which approved the gap on the author's behalf; default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203