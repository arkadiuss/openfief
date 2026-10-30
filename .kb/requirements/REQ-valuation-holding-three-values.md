---
title: A holding valuation at market must store, beside its gross value, the net and liquidation values the pack computes from its open lots' cost at the tax exchange rate, the account's wrapper, the sale cost under its fee rule and the owner's settings, or the reason each is missing, without failing the revaluation.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: A holding valuation at market must store, beside its gross value, the net and liquidation values the pack computes from its open lots' cost at the tax exchange rate, the account's wrapper, the sale cost under its fee rule and the owner's settings, or the reason each is missing, without failing the revaluation.
logic_claims:
  - CLAIM-87BDE8549C007264
semantic_clauses:
  - A holding valuation at market must store, beside its gross value, the net and liquidation values the pack computes from its open lots' cost at the tax exchange rate, the account's wrapper, the sale cost under its fee rule and the owner's settings, or the reason each is missing, without failing the revaluation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 417aca55407eee190c39a13e2e2656e84c54846d714034bcb24005962a97dec7
semantic_inventory:
  - claim_key: CLAIM-87BDE8549C007264
    claim_text: A holding valuation at market must store, beside its gross value, the net and liquidation values the pack computes from its open lots' cost at the tax exchange rate, the account's wrapper, the sale cost under its fee rule and the owner's settings, or the reason each is missing, without failing the revaluation
    role: normative
    status: modeled
    span:
      start: 0
      end: 310
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:54.176Z'
id: REQ-valuation-holding-three-values
type: req
---
A holding valuation at market must store, beside its gross value, the net and liquidation values the pack computes from its open lots' cost at the tax exchange rate, the account's wrapper, the sale cost under its fee rule and the owner's settings, or the reason each is missing, without failing the revaluation.

## Context

D-36 stores all three values of D-77 in each valuation. Converting each lot's cost at the NBP rate the pack takes for tax (D-20), the fee minimum at the day's rate, and storing a missing net or liquidation value with its reason are this package's readings. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-36]** Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

docs/DESIGN.md#D-36
