---
title: The core must serve the owner's ranking of market-data sources through a MarketDataService that only view plugins acting for an owner session may call, listing each kind's ranking and replacing one kind's order.
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: The core must serve the owner's ranking of market-data sources through a MarketDataService that only view plugins acting for an owner session may call, listing each kind's ranking and replacing one kind's order.
logic_claims:
  - CLAIM-5917F8D7691F8F65
semantic_clauses:
  - The core must serve the owner's ranking of market-data sources through a MarketDataService that only view plugins acting for an owner session may call, listing each kind's ranking and replacing one kind's order.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff644e34cb7a1aff1a0de957f15e5793dd1d47da4a58a0ced709ca05da558e00
semantic_inventory:
  - claim_key: CLAIM-5917F8D7691F8F65
    claim_text: The core must serve the owner's ranking of market-data sources through a MarketDataService that only view plugins acting for an owner session may call, listing each kind's ranking and replacing one kind's order
    role: normative
    status: modeled
    span:
      start: 0
      end: 210
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:22.369Z'
id: REQ-valuation-market-data-service
type: req
---
The core must serve the owner's ranking of market-data sources through a MarketDataService that only view plugins acting for an owner session may call, listing each kind's ranking and replacing one kind's order.

## Context

D-79 gives each kind of owner entry its own owner-only service; the ranking is one. Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94

> **[D-79]** What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69).

docs/DESIGN.md#D-79
