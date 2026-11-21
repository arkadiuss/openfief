---
title: The market price tool's card must show the instrument, its latest price in the owner's number format, the day of the close and the source, and the first, highest, lowest and latest close of the period, must list the instruments a ticker names with their ISINs and price none, and must hold the source's name as escaped data.
status: open
firmness: default
required_evidence: unverified
semantic_text: The market price tool's card must show the instrument, its latest price in the owner's number format, the day of the close and the source, and the first, highest, lowest and latest close of the period, must list the instruments a ticker names with their ISINs and price none, and must hold the source's name as escaped data.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-E11028D4BF85B2C7
    claim_text: The market price tool's card must show the instrument, its latest price in the owner's number format, the day of the close and the source, and the first, highest, lowest and latest close of the period, must list the instruments a ticker names with their ISINs and price none, and must hold the source's name as escaped data
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 323
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ask_box.market_price_card
logic_claims:
  - CLAIM-E11028D4BF85B2C7
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10ddfda7c03903f15d53c2a5de49f3032f667c66ebf90a9e7d8e82dd338da16c
rationale: 'D-226 asks for a price card and the author''s brief names its content: the instrument, the latest price, the source and the date; the four closes of the period, the list of candidates and the escaping are the implementer''s, for the author to review. The design system has no sparkline, so the card has none.'
id: REQ-ask-market-price-card
type: req
---
The market price tool's card must show the instrument, its latest price in the owner's number format, the day of the close and the source, and the first, highest, lowest and latest close of the period, must list the instruments a ticker names with their ISINs and price none, and must hold the source's name as escaped data.

## Context

D-226 asks for a price card and the author's brief names its content: the instrument, the latest price, the source and the date; the four closes of the period, the list of candidates and the escaping are the implementer's, for the author to review. The design system has no sparkline, so the card has none.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
