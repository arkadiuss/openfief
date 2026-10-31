---
title: Market index levels must not be ranked.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-4
  - lane:strict
text_ref: design-decisions:D-124#4
origin:
  kind: import
  ref: design-decisions:D-124#4
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Market index levels must not be ranked.
logic_claims:
  - CLAIM-D2AF021B667E7B19
semantic_clauses:
  - Market index levels must not be ranked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0fb701407b900fe17ad17a6fb26b1d194c7d1d4629f588e8031d546ab6bac07a
semantic_inventory:
  - claim_key: CLAIM-D2AF021B667E7B19
    claim_text: Market index levels must not be ranked
    role: normative
    status: modeled
    span:
      start: 0
      end: 38
id: REQ-valuation-index-levels-not-ranked
type: req
---
Market index levels must not be ranked.

## Context

D-94 ranks the sources of prices, exchange rates and index readings, where index readings are CPI and the NBP reference rate from the treasury bond connector (REQ-valuation-market-data-source-ranking). Market index levels, such as WIG20's, come from the market-data connector's index history (D-51, D-96) for the model portfolio's targets, and D-94 didn't say whether they are ranked. The implementation raised this as question #8 on docs/implementation-questions.md and keeps one level per index and date; on 2026-10-31 the author confirmed that market index levels aren't ranked, as D-124, and REQ-valuation-index-level-last-delivered says which level is kept.

## Source

> A source in D-94's ranking is one connector instance, ranked by default by the institution it reports, with the first-party connector of each kind first. Cross-check prices never rank: they only check (D-22). Market index levels aren't ranked; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#4
