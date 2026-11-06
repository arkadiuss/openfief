---
title: A question asked in the Ask box must appear in no log line of the UI or of the core's front door.
status: open
text_ref: docs/DESIGN.md#D-194
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-194
  - lane:strict
semantic_text: A question asked in the Ask box must appear in no log line of the UI or of the core's front door.
logic_claims:
  - CLAIM-F313108AB624D1D3
semantic_clauses:
  - A question asked in the Ask box must appear in no log line of the UI or of the core's front door.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 85b296fe63566743b209a4c55edfca832d46603c1d141ce9ee5d12be20ea0d91
semantic_inventory:
  - claim_key: CLAIM-F313108AB624D1D3
    claim_text: A question asked in the Ask box must appear in no log line of the UI or of the core's front door
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
origin:
  kind: agent
  recorded_at: '2026-11-06T08:18:17.362Z'
id: REQ-ui-ask-never-logged
type: req
---
A question asked in the Ask box must appear in no log line of the UI or of the core's front door.

## Context

D-194 and GR-12: the question never leaves the UI and is never logged. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-194]** In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date. Each answer is a card with a link to the place it summarizes, and anything else gets the commands it understands as suggestions. It never changes anything, the question never leaves the UI and is never logged (GR-12), and the box says it holds no AI model; in M2 an agent can take it over through the same answer cards (P-57).

docs/DESIGN.md#D-194
