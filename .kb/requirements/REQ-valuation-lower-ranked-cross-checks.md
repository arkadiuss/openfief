---
title: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
logic_claims:
  - CLAIM-C7E56B276B0BA899
semantic_clauses:
  - A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3fe9881365037ac24588e1f6ed9d79ca4f73493e811d1b91c746c1665fe4c404
semantic_inventory:
  - claim_key: CLAIM-C7E56B276B0BA899
    claim_text: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:42.609Z'
id: REQ-valuation-lower-ranked-cross-checks
type: req
---
A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.

## Context

D-94 makes the lower-ranked values the D-22 cross-checks. Cross-check prices stay apart because D-112 lets only a declared connector send them. Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
