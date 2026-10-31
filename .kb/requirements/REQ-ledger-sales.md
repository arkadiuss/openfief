---
title: The ledger must list an account's sells with their proceeds, fees and the lots each consumed, FIFO as lots are derived, and fail on an oversell.
status: open
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: The ledger must list an account's sells with their proceeds, fees and the lots each consumed, FIFO as lots are derived, and fail on an oversell.
logic_claims:
  - CLAIM-093CA50C90B86EDA
semantic_clauses:
  - The ledger must list an account's sells with their proceeds, fees and the lots each consumed, FIFO as lots are derived, and fail on an oversell.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4d49f12a6b5914dbede8e686fc01563b15624a159ceeb454124d757871ab8b4
semantic_inventory:
  - claim_key: CLAIM-093CA50C90B86EDA
    claim_text: The ledger must list an account's sells with their proceeds, fees and the lots each consumed, FIFO as lots are derived, and fail on an oversell
    role: normative
    status: modeled
    span:
      start: 0
      end: 143
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:09.518Z'
id: REQ-ledger-sales
type: req
---
The ledger must list an account's sells with their proceeds, fees and the lots each consumed, FIFO as lots are derived, and fail on an oversell.

## Context

A realized gain for D-128's totals needs each sale's consumed lots. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
