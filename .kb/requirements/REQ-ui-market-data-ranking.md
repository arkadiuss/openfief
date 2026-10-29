---
title: The built-in UI's settings page must show each kind's ranked sources with their institution and whether they are in their default place, and let the owner move a source up or down or restore the default order through the core's MarketDataService.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The built-in UI's settings page must show each kind's ranked sources with their institution and whether they are in their default place, and let the owner move a source up or down or restore the default order through the core's MarketDataService.
logic_claims:
  - CLAIM-6E5AF3167B7B46FA
semantic_clauses:
  - The built-in UI's settings page must show each kind's ranked sources with their institution and whether they are in their default place, and let the owner move a source up or down or restore the default order through the core's MarketDataService.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6e7272fe122bfd226924b00c09e5582e0e351f8dc9d2a6f23eae414294697f2e
semantic_inventory:
  - claim_key: CLAIM-6E5AF3167B7B46FA
    claim_text: The built-in UI's settings page must show each kind's ranked sources with their institution and whether they are in their default place, and let the owner move a source up or down or restore the default order through the core's MarketDataService
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:30.654Z'
id: REQ-ui-market-data-ranking
type: req
---
The built-in UI's settings page must show each kind's ranked sources with their institution and whether they are in their default place, and let the owner move a source up or down or restore the default order through the core's MarketDataService.

## Context

D-94 has the owner rank sources; the settings page is where the owner's settings live. Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
