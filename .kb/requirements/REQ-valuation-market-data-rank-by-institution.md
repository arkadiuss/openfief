---
title: By default, D-94's ranking must order connector instances by the institution each reports, with the first-party connector of each kind first.
status: closed
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-2
  - lane:strict
text_ref: design-decisions:D-124#2
origin:
  kind: import
  ref: design-decisions:D-124#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: By default, D-94's ranking must order connector instances by the institution each reports, with the first-party connector of each kind first.
logic_claims:
  - CLAIM-614F0D899188FA56
semantic_clauses:
  - By default, D-94's ranking must order connector instances by the institution each reports, with the first-party connector of each kind first.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 232363c396e21e35197208e1c18286bd685dc8d701cc54346f840d7ee795c3d9
semantic_inventory:
  - claim_key: CLAIM-614F0D899188FA56
    claim_text: By default, D-94's ranking must order connector instances by the institution each reports, with the first-party connector of each kind first
    role: normative
    status: modeled
    span:
      start: 0
      end: 140
id: REQ-valuation-market-data-rank-by-institution
type: req
---
By default, D-94's ranking must order connector instances by the institution each reports, with the first-party connector of each kind first.

## Context

D-94 ranks the first-party connector of each kind first by default: NBP for exchange rates, EODHD for prices and the treasury bond connector for CPI and the NBP reference rate (REQ-valuation-market-data-default-ranking). It didn't say how the core knows which instance is the first-party one. The implementation raised this as question #8 on docs/implementation-questions.md and ranks each instance by the institution it names on its deliveries, the kind's first-party institution first and the rest by name. On 2026-10-31 the author confirmed that as D-124; a source is one connector instance (REQ-valuation-market-data-source-instance).

## Source

> A source in D-94's ranking is one connector instance, ranked by default by the institution it reports, with the first-party connector of each kind first. Cross-check prices never rank: they only check (D-22). Market index levels aren't ranked; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#2
