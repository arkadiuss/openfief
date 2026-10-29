---
title: While the model portfolio is on, the built-in UI must draw its backtest as the value-over-time chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
status: open
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
semantic_text: While the model portfolio is on, the built-in UI must draw its backtest as the value-over-time chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
logic_claims:
  - CLAIM-166FFA1840224F73
semantic_clauses:
  - While the model portfolio is on, the built-in UI must draw its backtest as the value-over-time chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 01dd118678da26ef41acb265fc26b45e5d5ed2abb47be8aed4ebcfe6214c957a
semantic_inventory:
  - claim_key: CLAIM-166FFA1840224F73
    claim_text: While the model portfolio is on, the built-in UI must draw its backtest as the value-over-time chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core
    role: normative
    status: modeled
    span:
      start: 0
      end: 401
origin:
  kind: agent
  recorded_at: '2026-10-29T00:07:58.203Z'
id: REQ-ui-model-portfolio
type: req
---
While the model portfolio is on, the built-in UI must draw its backtest as the value-over-time chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.

## Context

D-52 (off and hidden by default, entered in the built-in UI and handed to the core), D-53 (second line from go-live, both at 100), D-54 (max drawdowns with dates). Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

docs/DESIGN.md#D-52

> **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100.

docs/DESIGN.md#D-53

> **[D-54]** M0 shows the historical max drawdown of the model and of each target: the worst fall from a peak to a later trough over its whole history, with both dates.

docs/DESIGN.md#D-54
