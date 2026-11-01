---
title: The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument.
logic_claims:
  - CLAIM-76DA85F5A44D246C
semantic_clauses:
  - The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6bb0a1d3505b29b0dc50229212c35910b714e1077fbe54fd16a3c36352907f09
semantic_inventory:
  - claim_key: CLAIM-76DA85F5A44D246C
    claim_text: The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument
    role: normative
    status: modeled
    span:
      start: 0
      end: 229
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:16.363Z'
id: REQ-valuation-market-data-series-service
type: req
---
The MarketDataService must let only a view plugin acting for an owner session list each series' source, change one series' source and set a kind's gap threshold, refusing an invalid series, source or threshold as invalid argument.

## Context

D-79 has one owner-only service per kind of entry; the series sources and gap thresholds of D-94 join the market-data rankings in MarketDataService, additively. Modeled by Claude Code; not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
