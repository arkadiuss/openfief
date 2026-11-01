---
title: Each instrument, currency pair and index must have exactly one source of its values, a connector instance.
status: open
text_ref: design-decisions:D-94#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-2
  - lane:strict
semantic_text: Each instrument, currency pair and index must have exactly one source of its values, a connector instance.
logic_claims:
  - CLAIM-C7E65AD611BD1700
semantic_clauses:
  - Each instrument, currency pair and index must have exactly one source of its values, a connector instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 90c4b882d50606855158ba382bf8b361f42558a554572b6efaee10715ad9b5de
semantic_inventory:
  - claim_key: CLAIM-C7E65AD611BD1700
    claim_text: Each instrument, currency pair and index must have exactly one source of its values, a connector instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#2
  recorded_at: '2026-11-01T07:36:47.497Z'
id: REQ-valuation-market-data-one-source
type: req
---
Each instrument, currency pair and index must have exactly one source of its values, a connector instance.

## Why this changed

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. REQ-valuation-market-data-highest-rank-wins let the highest-ranked source with a value win on each date, so a series could take its values from different sources on different days; this requirement replaces it.

## Context

D-94 governs where OpenFief's prices, exchange rates and index readings come from. D-124 says a source is one connector instance, and the other sources' values are kept only as cross-checks (D-22). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#2
