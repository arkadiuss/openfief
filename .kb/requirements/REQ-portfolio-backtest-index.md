---
title: The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even.
status: open
text_ref: docs/DESIGN.md#D-53
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-53
  - lane:strict
semantic_text: The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even.
logic_claims:
  - CLAIM-E707A47D88E7638E
semantic_clauses:
  - The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fcfb37800fc53bfe3fe9d458ac2012f0207c36345b2e2f27318c701c29d4390d
semantic_inventory:
  - claim_key: CLAIM-E707A47D88E7638E
    claim_text: The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even
    role: normative
    status: modeled
    span:
      start: 0
      end: 290
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:44.672Z'
id: REQ-portfolio-backtest-index
type: req
---
The backtest must start at 100 on the first day every target has history, hold the targets at their weights, rebalance to them at the end of each December 31 before its last day, end on the last day every target has history, and keep units to 18 and the index to 12 decimal places half-even.

## Context

D-53 says rebalanced once a year but not when; the calendar year end and the precision are this context's choices, parked for the author. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100.

docs/DESIGN.md#D-53
