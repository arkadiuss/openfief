---
title: 'Cross-check prices must never rank in D-94''s ranking: they only check.'
status: closed
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-3
  - lane:strict
text_ref: design-decisions:D-124#3
origin:
  kind: import
  ref: design-decisions:D-124#3
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: 'Cross-check prices must never rank in D-94''s ranking: they only check.'
logic_claims:
  - CLAIM-2646ABC4C5CC335B
semantic_clauses:
  - 'Cross-check prices must never rank in D-94''s ranking: they only check.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 49d2948e694bf9ef7fdcfc823b057539cb4e29808555433889e70d5c566216de
semantic_inventory:
  - claim_key: CLAIM-2646ABC4C5CC335B
    claim_text: 'Cross-check prices must never rank in D-94''s ranking: they only check'
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
id: REQ-valuation-cross-check-prices-never-rank
type: req
---
Cross-check prices must never rank in D-94's ranking: they only check.

## Context

D-22 keeps IBKR Flex prices as a cross-check of market-data prices (REQ-AUTO-7374F67D85269872), D-112 lets only a declared connector send them (REQ-ingest-cross-check-prices-declared-only), and D-94 has the owner rank the sources of each kind of market data, with lower-ranked values kept as cross-checks (REQ-valuation-market-data-all-sources-stored). D-94 didn't say whether cross-check prices are a ranked source. The implementation raised this as question #8 on docs/implementation-questions.md and stores them apart so they never win; on 2026-10-31 the author confirmed that as D-124, declining the option of ranking IBKR's mark prices as one more price source.

## Source

> A source in D-94's ranking is one connector instance, ranked by default by the institution it reports, with the first-party connector of each kind first. Cross-check prices never rank: they only check (D-22). Market index levels aren't ranked; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#3
