---
title: The model portfolio must be stored in the portfolio schema, read as off with no targets until first saved, and replaced whole, whether it is on and all its targets in the owner's order, in one database transaction.
status: open
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
semantic_text: The model portfolio must be stored in the portfolio schema, read as off with no targets until first saved, and replaced whole, whether it is on and all its targets in the owner's order, in one database transaction.
logic_claims:
  - CLAIM-A859FAF814B1E72D
semantic_clauses:
  - The model portfolio must be stored in the portfolio schema, read as off with no targets until first saved, and replaced whole, whether it is on and all its targets in the owner's order, in one database transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 81f14df4206ff807e38398cf41492003e089db91480210f2f663714f8110b7c4
semantic_inventory:
  - claim_key: CLAIM-A859FAF814B1E72D
    claim_text: The model portfolio must be stored in the portfolio schema, read as off with no targets until first saved, and replaced whole, whether it is on and all its targets in the owner's order, in one database transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 213
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:36.424Z'
id: REQ-portfolio-model-store
type: req
---
The model portfolio must be stored in the portfolio schema, read as off with no targets until first saved, and replaced whole, whether it is on and all its targets in the owner's order, in one database transaction.

## Context

D-52 makes the model optional and off by default; each context owns its schema. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

docs/DESIGN.md#D-52
