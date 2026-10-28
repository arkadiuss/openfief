---
title: A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets.
status: open
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
semantic_text: A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets.
logic_claims:
  - CLAIM-2D0B55914D301D47
semantic_clauses:
  - A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 48b784d59fa45af5c3a4a7f9ef6fc595f10a5473dee653278cd6fb52e30dcc75
semantic_inventory:
  - claim_key: CLAIM-2D0B55914D301D47
    claim_text: A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets
    role: normative
    status: modeled
    span:
      start: 0
      end: 277
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:35.720Z'
id: REQ-portfolio-model-targets
type: req
---
A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets.

## Context

D-52 fixes the kinds of target and the 100% sum. The 0.01% step, the 50-target limit, the index's currency and storing an off model without targets are this context's narrow choices so the owner can enter it in the UI; the author has not decided them. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

docs/DESIGN.md#D-52
