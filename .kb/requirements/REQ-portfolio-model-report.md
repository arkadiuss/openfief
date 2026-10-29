---
title: While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported.
status: open
text_ref: docs/DESIGN.md#D-53
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-53
  - lane:strict
semantic_text: While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported.
logic_claims:
  - CLAIM-015076512F752AD8
semantic_clauses:
  - While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 705c606f673ebbffe0c775f7227b81f128051a64031ddb045359b26f07a24191
semantic_inventory:
  - claim_key: CLAIM-015076512F752AD8
    claim_text: While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:41.072Z'
id: REQ-portfolio-model-report
type: req
---
While the model portfolio is on, its report must be its backtest from stored prices, index levels and FX rates through a date with the model's max drawdown and each target's; while it is off, nothing may be read or reported.

## Context

D-52 serves the model with its backtest and drawdowns only when on; D-53 and D-54 define them. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100.

docs/DESIGN.md#D-53

> **[D-54]** M0 shows the historical max drawdown of the model and of each target: the worst fall from a peak to a later trough over its whole history, with both dates.

docs/DESIGN.md#D-54
