---
title: The max drawdown of a series must be the largest fall from a day's value to a later day's value as a fraction of the first, with both dates and values, keeping the earliest of equal falls, rounded half-even to 12 places, and zero with the first day as peak and trough when the series never falls.
status: open
text_ref: docs/DESIGN.md#D-54
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-54
  - lane:strict
semantic_text: The max drawdown of a series must be the largest fall from a day's value to a later day's value as a fraction of the first, with both dates and values, keeping the earliest of equal falls, rounded half-even to 12 places, and zero with the first day as peak and trough when the series never falls.
logic_claims:
  - CLAIM-6B758EE89AFF92F1
semantic_clauses:
  - The max drawdown of a series must be the largest fall from a day's value to a later day's value as a fraction of the first, with both dates and values, keeping the earliest of equal falls, rounded half-even to 12 places, and zero with the first day as peak and trough when the series never falls.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a430c56cf8db02b11c6253acfd3c494a831b8ee441412ad68259a6b798bbaff
semantic_inventory:
  - claim_key: CLAIM-6B758EE89AFF92F1
    claim_text: The max drawdown of a series must be the largest fall from a day's value to a later day's value as a fraction of the first, with both dates and values, keeping the earliest of equal falls, rounded half-even to 12 places, and zero with the first day as peak and trough when the series never falls
    role: normative
    status: modeled
    span:
      start: 0
      end: 295
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:49.118Z'
id: REQ-portfolio-drawdown-depth
type: req
---
The max drawdown of a series must be the largest fall from a day's value to a later day's value as a fraction of the first, with both dates and values, keeping the earliest of equal falls, rounded half-even to 12 places, and zero with the first day as peak and trough when the series never falls.

## Context

D-54 asks for the worst fall from a peak to a later trough with both dates. Ties and a series that never falls are this context's reading. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-54]** M0 shows the historical max drawdown of the model and of each target: the worst fall from a peak to a later trough over its whole history, with both dates.

docs/DESIGN.md#D-54
