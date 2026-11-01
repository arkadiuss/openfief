---
title: A series' source must be chosen from the owner's ranking of sources for that kind of data when the series first needs a value.
status: open
text_ref: design-decisions:D-94#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-5
  - lane:strict
semantic_text: A series' source must be chosen from the owner's ranking of sources for that kind of data when the series first needs a value.
logic_claims:
  - CLAIM-47C7D4D1AC33FDDF
semantic_clauses:
  - A series' source must be chosen from the owner's ranking of sources for that kind of data when the series first needs a value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4c13c01b16e0b535bcde452d1bee2ab38031fca940e8101d60a7e522fa02d192
semantic_inventory:
  - claim_key: CLAIM-47C7D4D1AC33FDDF
    claim_text: A series' source must be chosen from the owner's ranking of sources for that kind of data when the series first needs a value
    role: normative
    status: modeled
    span:
      start: 0
      end: 125
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#5
  recorded_at: '2026-11-01T07:37:02.165Z'
id: REQ-valuation-market-data-source-chosen-on-first-need
type: req
---
A series' source must be chosen from the owner's ranking of sources for that kind of data when the series first needs a value.

## Why this changed

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. The ranking no longer decides each date's value; it picks a series' one source when that series first needs a value. The rewritten D-124 also drops its default ordering of instances by the institution each reports, which REQ-valuation-market-data-rank-by-institution held, so this requirement replaces it; the first-party connector of each kind still ranks first by default (REQ-valuation-market-data-default-ranking).

## Context

A series is an instrument, a currency pair or an index (D-94). The owner ranks the sources of each kind of market data (REQ-valuation-market-data-source-ranking). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#5
