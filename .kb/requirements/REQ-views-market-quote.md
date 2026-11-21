---
title: The read API's GetMarketQuote must look an instrument up by ticker or ISIN and return its latest end-of-day price, the closes of a period of at most 366 days, 30 by default, the connector instance that answered and the day of the latest close, as data tier T2, and must answer a ticker that names several instruments with them and no price.
status: open
firmness: default
required_evidence: unverified
semantic_text: The read API's GetMarketQuote must look an instrument up by ticker or ISIN and return its latest end-of-day price, the closes of a period of at most 366 days, 30 by default, the connector instance that answered and the day of the latest close, as data tier T2, and must answer a ticker that names several instruments with them and no price.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-850E90D4B42A5571
    claim_text: The read API's GetMarketQuote must look an instrument up by ticker or ISIN and return its latest end-of-day price, the closes of a period of at most 366 days, 30 by default, the connector instance that answered and the day of the latest close, as data tier T2, and must answer a ticker that names several instruments with them and no price
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 339
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - views.market_quote
logic_claims:
  - CLAIM-850E90D4B42A5571
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 417e07332c8717500415cb2202578198122de2a6871a2d4ddc090959fe4ee52d
rationale: D-226 lets the Ask box's model fetch prices for any instrument through the owner's market-data connector, and D-6 and D-213 make the read API the one place the Ask box and MCP read from. The author decided it on 2026-11-21; the shape of the answer, its period and its candidates are the implementer's, for the author to review.
id: REQ-views-market-quote
type: req
---
The read API's GetMarketQuote must look an instrument up by ticker or ISIN and return its latest end-of-day price, the closes of a period of at most 366 days, 30 by default, the connector instance that answered and the day of the latest close, as data tier T2, and must answer a ticker that names several instruments with them and no price.

## Context

D-226 lets the Ask box's model fetch prices for any instrument through the owner's market-data connector, and D-6 and D-213 make the read API the one place the Ask box and MCP read from. The author decided it on 2026-11-21; the shape of the answer, its period and its candidates are the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
