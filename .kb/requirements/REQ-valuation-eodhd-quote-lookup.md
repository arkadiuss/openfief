---
title: The EODHD connector's LookupQuote must find the one listing a ticker or ISIN names on any exchange, preferring a US venue and then the exchanges of D-51 for one ISIN listed on several, must answer several ISINs under one ticker with at most ten candidates and no quote, must fetch only through its egress grant, and must answer a period with no close with the latest close before it.
status: open
firmness: default
required_evidence: unverified
semantic_text: The EODHD connector's LookupQuote must find the one listing a ticker or ISIN names on any exchange, preferring a US venue and then the exchanges of D-51 for one ISIN listed on several, must answer several ISINs under one ticker with at most ten candidates and no quote, must fetch only through its egress grant, and must answer a period with no close with the latest close before it.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-DDFEFDD3DB45CF15
    claim_text: The EODHD connector's LookupQuote must find the one listing a ticker or ISIN names on any exchange, preferring a US venue
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 121
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - valuation.eodhd_quote_lookup
  - claim_key: CLAIM-92826B613E05EFFF
    claim_text: then the exchanges of D-51 for one ISIN listed on several, must answer several ISINs under one ticker with at most ten candidates
    role: normative
    status: ontology_gap
    span:
      start: 126
      end: 255
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - valuation.eodhd_quote_lookup
  - claim_key: CLAIM-98EA3AA7CAFA80C4
    claim_text: no quote, must fetch only through its egress grant, and must answer a period with no close with the latest close before it
    role: normative
    status: ontology_gap
    span:
      start: 260
      end: 382
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - valuation.eodhd_quote_lookup
logic_claims:
  - CLAIM-DDFEFDD3DB45CF15
  - CLAIM-92826B613E05EFFF
  - CLAIM-98EA3AA7CAFA80C4
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d44f37831ae5ac4d3f8847f198c295c6c5ff26f61b86142eb1385c04f99f9bcc
rationale: D-226 names the owner's market-data connector as the source, and D-51 makes EODHD the first one; EODHD lists one instrument on many venues, so the implementer chose how a lookup picks one, with the US venue first because the question that started it names NVDA. The choice is the implementer's, for the author to review.
id: REQ-valuation-eodhd-quote-lookup
type: req
---
The EODHD connector's LookupQuote must find the one listing a ticker or ISIN names on any exchange, preferring a US venue and then the exchanges of D-51 for one ISIN listed on several, must answer several ISINs under one ticker with at most ten candidates and no quote, must fetch only through its egress grant, and must answer a period with no close with the latest close before it.

## Context

D-226 names the owner's market-data connector as the source, and D-51 makes EODHD the first one; EODHD lists one instrument on many venues, so the implementer chose how a lookup picks one, with the US venue first because the question that started it names NVDA. The choice is the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1
