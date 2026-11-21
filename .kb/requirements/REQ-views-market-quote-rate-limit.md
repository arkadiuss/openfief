---
title: The views must limit the market quote lookups of each caller to 10 a minute and of all callers together to 60 a minute, and must refuse a lookup over either limit as resource exhausted without asking the market-data connector.
status: open
firmness: default
required_evidence: unverified
semantic_text: The views must limit the market quote lookups of each caller to 10 a minute and of all callers together to 60 a minute, and must refuse a lookup over either limit as resource exhausted without asking the market-data connector.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-6086EAD05C734D1A
    claim_text: The views must limit the market quote lookups of each caller to 10 a minute
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 75
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - views.market_quote_rate_limit
  - claim_key: CLAIM-B8AA22DA52111025
    claim_text: of all callers together to 60 a minute, and must refuse a lookup over either limit as resource exhausted without asking the market-data connector
    role: normative
    status: ontology_gap
    span:
      start: 80
      end: 225
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - views.market_quote_rate_limit
logic_claims:
  - CLAIM-6086EAD05C734D1A
  - CLAIM-B8AA22DA52111025
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2b5af42cdb685a9db72cb6cab2a2d406839e82de4cf7f6504d7dfc1dfe38d233
rationale: 'Inferred from the read API''s script tokens, which have no rate limit, and from D-226: a looping token or a model would spend the owner''s provider key, and one caller''s loop must not starve the others. No one asked for the limits; their numbers are the implementer''s, for the author to review.'
id: REQ-views-market-quote-rate-limit
type: req
---
The views must limit the market quote lookups of each caller to 10 a minute and of all callers together to 60 a minute, and must refuse a lookup over either limit as resource exhausted without asking the market-data connector.

## Context

Inferred from the read API's script tokens, which have no rate limit, and from D-226: a looping token or a model would spend the owner's provider key, and one caller's loop must not starve the others. No one asked for the limits; their numbers are the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
