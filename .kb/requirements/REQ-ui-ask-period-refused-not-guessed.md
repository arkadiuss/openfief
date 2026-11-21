---
title: The Ask box must answer a period that ends before it starts, a period that ends after the latest day an account is valued through and a period the core refuses with a card that says why, and must never shorten, reorder or guess the period.
status: open
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: the Ask box performance and value tools and commands answer for a bounded period, a gap against D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T13:36:20.162Z'
firmness: default
text_ref: docs/DESIGN.md#D-194
semantic_text: The Ask box must answer a period that ends before it starts, a period that ends after the latest day an account is valued through and a period the core refuses with a card that says why, and must never shorten, reorder or guess the period.
tags:
  - ui
  - ask
  - period
  - d-194
  - d-203
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a96306d81219af3bc5268dd7824511e1f8061b49b0ab450e40d3bdebd062820d
semantic_inventory:
  - claim_key: CLAIM-F8B03EA95E075589
    claim_text: The Ask box must answer a period that ends before it starts, a period that ends after the latest day an account is valued through
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 129
    subjects:
      - ask_box.period_refusals
  - claim_key: CLAIM-AFF1220A48715348
    claim_text: a period the core refuses with a card that says why, and must never shorten, reorder or guess the period
    role: normative
    status: ontology_gap
    span:
      start: 134
      end: 238
    subjects:
      - ask_box.period_refusals
logic_claims:
  - CLAIM-F8B03EA95E075589
  - CLAIM-AFF1220A48715348
id: REQ-ui-ask-period-refused-not-guessed
type: req
---
The Ask box must answer a period that ends before it starts, a period that ends after the latest day an account is valued through and a period the core refuses with a card that says why, and must never shorten, reorder or guess the period.

## Context

D-194 lists performance over a period and the value on a date as commands of the Ask box, and D-203 lets its model call only the same read-only answerers. Before this, only periods that start on a fixed day (this year, since a date, a calendar year) could be asked, so a question about one month, such as why the value or the return changed in November 2025, could not be answered. The read API answers a bounded period first (REQ-views-period-change), and the box shows its figures. Recorded by the implementation agent for the design session, which approved the gap on the author's behalf; default firmness.

## Source

> In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date.

docs/DESIGN.md#D-194, D-203