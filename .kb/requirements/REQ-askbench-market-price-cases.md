---
title: 'The Ask box bench must hold cases for market prices over a fake market-data connector of made-up instruments: NVDA''s price in English and Polish, an unknown ticker that gets no price from memory, a route that may not receive T2 whose answer says prices need a model cleared for market data, and an instruction planted in an instrument name the connector returns.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The Ask box bench must hold cases for market prices over a fake market-data connector of made-up instruments: NVDA''s price in English and Polish, an unknown ticker that gets no price from memory, a route that may not receive T2 whose answer says prices need a model cleared for market data, and an instruction planted in an instrument name the connector returns.'
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-E3133A15007E5739
    claim_text: 'The Ask box bench must hold cases for market prices over a fake market-data connector of made-up instruments: NVDA''s price in English and Polish, an unknown ticker that gets no price from memory, a route that may not receive T2 whose answer says prices need a model cleared for market data, and an instruction planted in an instrument name the connector returns'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 361
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ask_bench.market_price_cases
logic_claims:
  - CLAIM-E3133A15007E5739
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fc67b1d67010cba20767dcb71cb86b646e869f105ad0b405c923fa52d135c933
rationale: D-226 and the brief for this change ask for these four kinds of case; D-219 puts injections planted in names into the bench's seed variant. The ids P1 to P5 and the made-up closes are the implementer's, and P4 and P5 are safety cases, for the author to review.
id: REQ-askbench-market-price-cases
type: req
---
The Ask box bench must hold cases for market prices over a fake market-data connector of made-up instruments: NVDA's price in English and Polish, an unknown ticker that gets no price from memory, a route that may not receive T2 whose answer says prices need a model cleared for market data, and an instruction planted in an instrument name the connector returns.

## Context

D-226 and the brief for this change ask for these four kinds of case; D-219 puts injections planted in names into the bench's seed variant. The ids P1 to P5 and the made-up closes are the implementer's, and P4 and P5 are safety cases, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
