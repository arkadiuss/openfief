---
title: While the model portfolio is on, the built-in UI must draw its backtest as the performance chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
status: open
text_ref: docs/DESIGN.md#D-106
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-106
  - lane:strict
semantic_text: While the model portfolio is on, the built-in UI must draw its backtest as the performance chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
logic_claims:
  - CLAIM-B450822A1A83B493
semantic_clauses:
  - While the model portfolio is on, the built-in UI must draw its backtest as the performance chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 441e596471c33ba792875f6fbd6e6816959c6500ccaf75987bd2a492033f2424
semantic_inventory:
  - claim_key: CLAIM-B450822A1A83B493
    claim_text: While the model portfolio is on, the built-in UI must draw its backtest as the performance chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core
    role: normative
    status: modeled
    span:
      start: 0
      end: 397
origin:
  kind: agent
  recorded_at: '2026-10-30T03:39:40.976Z'
id: REQ-ui-model-portfolio-performance-chart
type: req
---
While the model portfolio is on, the built-in UI must draw its backtest as the performance chart's second line on the time-weighted return's days, both 100 on go-live, and list the model's and each target's max drawdown with its dates; it must show nothing of it while it is off; and its settings page must let only an owner session turn it on or off and edit its targets, handing them to the core.

## Context

D-53 and D-54 compare the portfolio with the model and its drawdowns; D-106 puts the model next to the time-weighted return on the performance chart, apart from the value chart in money. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-ui-model-portfolio: D-106 moves the model's line from the value-over-time chart to the performance chart. The superseded requirement is closed.

## Source

> **[D-106]** Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

docs/DESIGN.md#D-106
