---
title: The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-1
  - lane:strict
text_ref: design-decisions:D-94#1
origin:
  kind: import
  ref: design-decisions:D-94#1
  recorded_at: '2026-10-29T08:48:52.652Z'
semantic_text: The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order.
logic_claims:
  - CLAIM-E8238FC7EE4C24F7
semantic_clauses:
  - The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 16eed8bafc3c96a46e0933a155e418aec5bfc01b05c8b8ef1cb586ad65f95cea
semantic_inventory:
  - claim_key: CLAIM-E8238FC7EE4C24F7
    claim_text: The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order
    role: normative
    status: modeled
    span:
      start: 0
      end: 121
id: REQ-valuation-market-data-source-ranking
type: req
---
The owner must rank the sources of each kind of market data, prices, exchange rates and index readings, in priority order.

## Context

Several connectors can deliver the same kind of market data: D-22 keeps IBKR Flex prices as a cross-check of market-data prices (REQ-AUTO-7374F67D85269872), and D-20 makes the FX rate provider swappable (REQ-AUTO-D1568676DF81CAF7). D-94 replaces letting the last write win: the owner ranks the sources of each kind, the highest-ranked source with a value wins for each date (REQ-valuation-market-data-highest-rank-wins), every source's values are kept (REQ-valuation-market-data-all-sources-stored), and the first-party connectors rank first by default (REQ-valuation-market-data-default-ranking). The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#1
