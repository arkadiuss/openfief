---
title: A series rebased on a date must start at 100 on its first day on or after that date, each value multiplied by 100 and divided by that day's value, rounded half-even to 12 decimal places.
status: open
text_ref: docs/DESIGN.md#D-53
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-53
  - lane:strict
semantic_text: A series rebased on a date must start at 100 on its first day on or after that date, each value multiplied by 100 and divided by that day's value, rounded half-even to 12 decimal places.
logic_claims:
  - CLAIM-0A572D01C3C97709
semantic_clauses:
  - A series rebased on a date must start at 100 on its first day on or after that date, each value multiplied by 100 and divided by that day's value, rounded half-even to 12 decimal places.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ea559c187fb3f913938555202018ed3d70c57966781c1e2fa6d7521b181da02b
semantic_inventory:
  - claim_key: CLAIM-0A572D01C3C97709
    claim_text: A series rebased on a date must start at 100 on its first day on or after that date, each value multiplied by 100 and divided by that day's value, rounded half-even to 12 decimal places
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:53.577Z'
id: REQ-portfolio-rebase
type: req
---
A series rebased on a date must start at 100 on its first day on or after that date, each value multiplied by 100 and divided by that day's value, rounded half-even to 12 decimal places.

## Context

D-53 starts both chart lines at 100 on go-live; the model's backtest usually starts earlier, so its line is rebased. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100.

docs/DESIGN.md#D-53
