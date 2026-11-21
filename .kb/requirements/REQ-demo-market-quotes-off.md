---
title: 'The demo must answer from its synthetic data alone: with demo sessions the core must give the views no market quote source, GetMarketQuote must be a failed precondition, GetSettings must say market quotes are not available and the Ask box must not offer the market price tool.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The demo must answer from its synthetic data alone: with demo sessions the core must give the views no market quote source, GetMarketQuote must be a failed precondition, GetSettings must say market quotes are not available and the Ask box must not offer the market price tool.'
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-4870E0C5FD1780DA
    claim_text: 'The demo must answer from its synthetic data alone: with demo sessions the core must give the views no market quote source, GetMarketQuote must be a failed precondition, GetSettings must say market quotes are not available'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 222
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - demo.market_quotes
  - claim_key: CLAIM-0DF9D803D5547DE6
    claim_text: the Ask box must not offer the market price tool
    role: normative
    status: ontology_gap
    span:
      start: 227
      end: 275
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - demo.market_quotes
logic_claims:
  - CLAIM-4870E0C5FD1780DA
  - CLAIM-0DF9D803D5547DE6
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6b9052a5e2d8a1006d6ba6d4e8ae456a1dd597f43c87d7c9979c0d77f24203c3
rationale: D-226 says the demo answers from its synthetic data alone, and the brief for this change leaves the choice between synthetic quotes and no tool to the implementer, who chose no tool because the demo installs no market-data connector with a key and show already gives a held instrument's stored price. The choice is for the author to review.
id: REQ-demo-market-quotes-off
type: req
---
The demo must answer from its synthetic data alone: with demo sessions the core must give the views no market quote source, GetMarketQuote must be a failed precondition, GetSettings must say market quotes are not available and the Ask box must not offer the market price tool.

## Context

D-226 says the demo answers from its synthetic data alone, and the brief for this change leaves the choice between synthetic quotes and no tool to the implementer, who chose no tool because the demo installs no market-data connector with a key and show already gives a held instrument's stored price. The choice is for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
