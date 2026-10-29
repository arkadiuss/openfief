---
title: 'By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-4
  - lane:strict
text_ref: design-decisions:D-94#4
origin:
  kind: import
  ref: design-decisions:D-94#4
  recorded_at: '2026-10-29T09:39:39.959Z'
semantic_text: 'By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.'
logic_claims:
  - CLAIM-8D1DCBFD1A770177
semantic_clauses:
  - 'By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b069e21ee9491081c509bc3689fe4e87a6f9a33f37e345e9867e1b06d51deead
semantic_inventory:
  - claim_key: CLAIM-8D1DCBFD1A770177
    claim_text: 'By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate'
    role: normative
    status: modeled
    span:
      start: 0
      end: 191
id: REQ-valuation-market-data-default-ranking
type: req
---
By default the first-party connector for each kind of market data must rank first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

## Context

D-94 has the owner rank the sources of each kind of market data (REQ-valuation-market-data-source-ranking). Until the owner changes it, the first-party connector ranks first: NBP, the default FX data plugin of D-20 (REQ-AUTO-D1568676DF81CAF7), for exchange rates; EODHD, the first market-data connector of D-51 (REQ-valuation-eodhd-connector), for prices; and the treasury bond connector for CPI and the NBP reference rate. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#4
