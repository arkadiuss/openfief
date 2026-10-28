---
title: The read API must answer not found and name nothing about the model portfolio while it is off, and while it is on serve its targets, its backtest rebased to 100 on the first day from go-live or the asked date through the latest valued-through date or the asked date, the backtest's bounds and the max drawdowns of the model and of each target.
status: open
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
semantic_text: The read API must answer not found and name nothing about the model portfolio while it is off, and while it is on serve its targets, its backtest rebased to 100 on the first day from go-live or the asked date through the latest valued-through date or the asked date, the backtest's bounds and the max drawdowns of the model and of each target.
logic_claims:
  - CLAIM-85057294B627D0F5
semantic_clauses:
  - The read API must answer not found and name nothing about the model portfolio while it is off, and while it is on serve its targets, its backtest rebased to 100 on the first day from go-live or the asked date through the latest valued-through date or the asked date, the backtest's bounds and the max drawdowns of the model and of each target.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 15cb1fe53fcfb8e41f01c5618be723bbd9211a788bfb2c6b756a04c8924f8420
semantic_inventory:
  - claim_key: CLAIM-85057294B627D0F5
    claim_text: The read API must answer not found and name nothing about the model portfolio while it is off, and while it is on serve its targets, its backtest rebased to 100 on the first day from go-live or the asked date through the latest valued-through date or the asked date, the backtest's bounds and the max drawdowns of the model and of each target
    role: normative
    status: modeled
    span:
      start: 0
      end: 342
origin:
  kind: agent
  recorded_at: '2026-10-28T23:32:32.803Z'
id: REQ-views-model-portfolio
type: req
---
The read API must answer not found and name nothing about the model portfolio while it is off, and while it is on serve its targets, its backtest rebased to 100 on the first day from go-live or the asked date through the latest valued-through date or the asked date, the backtest's bounds and the max drawdowns of the model and of each target.

## Context

D-52 serves the model with its backtest and drawdowns only when on; D-53 starts both chart lines at 100 on go-live. Answering 404 while off, rather than an empty answer, is this context's reading. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

docs/DESIGN.md#D-52

> **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100.

docs/DESIGN.md#D-53
