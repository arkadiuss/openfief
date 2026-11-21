---
title: The Ask box must offer its market price tool to a model only where a market-data connector runs and the router allows a model of the route to receive the question's T0 and the price's T2 together, must tag the price card T2 and withhold it when the route no longer allows T2, and must make its system prompt say that prices need a model cleared for market data where the route does not.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box must offer its market price tool to a model only where a market-data connector runs and the router allows a model of the route to receive the question's T0 and the price's T2 together, must tag the price card T2 and withhold it when the route no longer allows T2, and must make its system prompt say that prices need a model cleared for market data where the route does not.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-D3CCA3B0E7268A0F
    claim_text: The Ask box must offer its market price tool to a model only where a market-data connector runs and the router allows a model of the route to receive the question's T0 and the price's T2 together, must tag the price card T2 and withhold it when the route no longer allows T2, and must make its system prompt say that prices need a model cleared for market data where the route does not
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 385
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ask_box.market_price_offer
logic_claims:
  - CLAIM-D3CCA3B0E7268A0F
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb270df2e411faac8f0704298e3248fee95ab6af645d28d1b805159df6ff0c9c
rationale: D-226 says the prices are T2 data that only a model allowed for T2 receives, and D-206 has the router send a request only to a model allowed for every tier in it. The author decided those on 2026-11-21 and earlier; the check before the tool is offered and the sentence in the prompt are the implementer's, for the author to review.
id: REQ-ask-market-price-offer
type: req
---
The Ask box must offer its market price tool to a model only where a market-data connector runs and the router allows a model of the route to receive the question's T0 and the price's T2 together, must tag the price card T2 and withhold it when the route no longer allows T2, and must make its system prompt say that prices need a model cleared for market data where the route does not.

## Context

D-226 says the prices are T2 data that only a model allowed for T2 receives, and D-206 has the router send a request only to a model allowed for every tier in it. The author decided those on 2026-11-21 and earlier; the check before the tool is offered and the sentence in the prompt are the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
