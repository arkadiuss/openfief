---
title: A source in D-94's ranking of market data must be one connector instance.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-1
  - lane:strict
text_ref: design-decisions:D-124#1
origin:
  kind: import
  ref: design-decisions:D-124#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: A source in D-94's ranking of market data must be one connector instance.
logic_claims:
  - CLAIM-215E4C2DF0855B46
semantic_clauses:
  - A source in D-94's ranking of market data must be one connector instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 94813b258cf9514c605d02276e76c4d95172fc0fa1467a0e55e599fd8069ae78
semantic_inventory:
  - claim_key: CLAIM-215E4C2DF0855B46
    claim_text: A source in D-94's ranking of market data must be one connector instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 72
id: REQ-valuation-market-data-source-instance
type: req
---
A source in D-94's ranking of market data must be one connector instance.

## Context

D-94 has the owner rank the sources of each kind of market data (REQ-valuation-market-data-source-ranking), but didn't say whether a source is a connector instance or a plugin. The implementation raised this as question #8 on docs/implementation-questions.md and built the narrowest reading: a source is the connector instance that delivered the value, as D-48 names instances. On 2026-10-31 the author confirmed that reading as D-124, so two instances of one plugin are two sources the owner can rank apart.

## Source

> A source in D-94's ranking is one connector instance, ranked by default by the institution it reports, with the first-party connector of each kind first. Cross-check prices never rank: they only check (D-22). Market index levels aren't ranked; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#1
