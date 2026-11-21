---
title: 'The core must take a market-data connector''s quote only as data: a query of at most 32 letters, digits, dots, dashes and underscores, a period of at most 366 days ending today or earlier, a listing text cut to 200 characters on one line of printable text, and closes that are plain positive amounts in the listing''s currency, in date order and no later than the period''s last day.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The core must take a market-data connector''s quote only as data: a query of at most 32 letters, digits, dots, dashes and underscores, a period of at most 366 days ending today or earlier, a listing text cut to 200 characters on one line of printable text, and closes that are plain positive amounts in the listing''s currency, in date order and no later than the period''s last day.'
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-9CA71FF050E9048E
    claim_text: 'The core must take a market-data connector''s quote only as data: a query of at most 32 letters, digits, dots, dashes and underscores, a period of at most 366 days ending today or earlier, a listing text cut to 200 characters on one line of printable text, and closes that are plain positive amounts in the listing''s currency, in date order and no later than the period''s last day'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 379
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ingest.market_quote_answer
logic_claims:
  - CLAIM-9CA71FF050E9048E
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64421593003dc1734d5fcc65c38c0197064c3973b428ca3015042fffc083ed04
rationale: R-6 and D-219 treat text a stored record or a source carries as data that may try to instruct the Ask box's model, and a provider's instrument name can carry any words. The implementer applied that to the connector's quote; the limits are the implementer's, for the author to review.
id: REQ-ingest-market-quote-answer-is-data
type: req
---
The core must take a market-data connector's quote only as data: a query of at most 32 letters, digits, dots, dashes and underscores, a period of at most 366 days ending today or earlier, a listing text cut to 200 characters on one line of printable text, and closes that are plain positive amounts in the listing's currency, in date order and no later than the period's last day.

## Context

R-6 and D-219 treat text a stored record or a source carries as data that may try to instruct the Ask box's model, and a provider's instrument name can carry any words. The implementer applied that to the connector's quote; the limits are the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
