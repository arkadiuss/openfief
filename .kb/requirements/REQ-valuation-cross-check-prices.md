---
title: Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation.
logic_claims:
  - CLAIM-C17FFBB9B4B2F0DD
semantic_clauses:
  - Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4d48573a60bef40c6bea8397bb6c2c4f6ece55cb35f696961ef4e98a427de94a
semantic_inventory:
  - claim_key: CLAIM-C17FFBB9B4B2F0DD
    claim_text: Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation
    role: normative
    status: modeled
    span:
      start: 0
      end: 253
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:29.209Z'
id: REQ-valuation-cross-check-prices
type: req
---
Cross-check prices must be stored beside the valuation prices, keyed by instrument, date and source, each with an instrument, a date, a positive close and a source, must never replace a valuation price or value a holding, and must enqueue no revaluation.

## Context

D-22 makes the market-data connector the source of prices and IBKR's prices a cross-check only, so the two must not share a table: a later write replaces a price (implementation questions, shared market data, undecided), and IBKR's marks there would change valuations. Keying by source keeps each IBKR instance's reading (D-48). Refines REQ-AUTO-7374F67D85269872 and REQ-AUTO-0C514DFCDB36FBEB. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22
