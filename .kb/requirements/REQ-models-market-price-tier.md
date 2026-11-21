---
title: A market price in a conversation must be tagged T2 and the Ask box's own fixed text, its system prompt, its retry and final-round texts and its messages that a tool failed or a card was withheld, must be tagged instructions, which adds no tier to a request, so that a route allowed T0 only answers a question about the portfolio and declines a price.
status: open
firmness: default
required_evidence: unverified
semantic_text: A market price in a conversation must be tagged T2 and the Ask box's own fixed text, its system prompt, its retry and final-round texts and its messages that a tool failed or a card was withheld, must be tagged instructions, which adds no tier to a request, so that a route allowed T0 only answers a question about the portfolio and declines a price.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
text_ref: design-decisions:D-226#1
rationale: D-206 tiers every message of a request and D-226 makes a price T2; tagging the box's fixed text T2 too would refuse a T0-only route every question, so the design session ruled that fixed text carries no tier. How the tag is spelled, instructions, is the implementer's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7f8ae36bb0efa8113a67c93ddbf7cc10ccd3f6e242e31d2f2f0004076c97c1d0
semantic_inventory:
  - claim_key: CLAIM-B7D3C747668A47ED
    claim_text: A market price in a conversation must be tagged T2 and the Ask box's own fixed text, its system prompt, its retry and final-round texts and its messages that a tool failed or a card was withheld, must be tagged instructions, which adds no tier to a request, so that a route allowed T0 only answers a question about the portfolio and declines a price
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 349
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - models.market_price_tier
logic_claims:
  - CLAIM-B7D3C747668A47ED
id: REQ-models-market-price-tier
type: req
---
A market price in a conversation must be tagged T2 and the Ask box's own fixed text, its system prompt, its retry and final-round texts and its messages that a tool failed or a card was withheld, must be tagged instructions, which adds no tier to a request, so that a route allowed T0 only answers a question about the portfolio and declines a price.

## Context

D-206 tiers every message of a request and D-226 makes a price T2; tagging the box's fixed text T2 too would refuse a T0-only route every question, so the design session ruled that fixed text carries no tier. How the tag is spelled, instructions, is the implementer's.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
