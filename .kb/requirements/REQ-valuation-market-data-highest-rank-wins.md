---
title: For each kind of market data and each date, the highest-ranked source with a value must win.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-2
  - lane:strict
text_ref: design-decisions:D-94#2
origin:
  kind: import
  ref: design-decisions:D-94#2
  recorded_at: '2026-10-29T08:49:29.421Z'
semantic_text: For each kind of market data and each date, the highest-ranked source with a value must win.
logic_claims:
  - CLAIM-0345FA5C43BC4433
semantic_clauses:
  - For each kind of market data and each date, the highest-ranked source with a value must win.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7e621d24f0d6df78052a91845697e88cfd78d4296e37c4e3c9c2bcaca9c6ea1c
semantic_inventory:
  - claim_key: CLAIM-0345FA5C43BC4433
    claim_text: For each kind of market data and each date, the highest-ranked source with a value must win
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
id: REQ-valuation-market-data-highest-rank-wins
type: req
---
For each kind of market data and each date, the highest-ranked source with a value must win.

## Context

D-94 replaces letting the last write win for market data. The owner ranks the sources of each kind (REQ-valuation-market-data-source-ranking), and on each date the value used is the one from the highest-ranked source that has a value for that date, so a lower-ranked source fills a gap without ever overriding a better one. REQ-valuation-fx-rate-lookup and REQ-valuation-index-levels, which the author has not reviewed, still treat two rates for one date as an error and let a later index level replace an earlier one whatever its source. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#2
