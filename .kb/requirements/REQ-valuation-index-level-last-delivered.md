---
title: Each market index and date must keep the level delivered last.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-5
  - lane:strict
text_ref: design-decisions:D-124#5
origin:
  kind: import
  ref: design-decisions:D-124#5
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Each market index and date must keep the level delivered last.
logic_claims:
  - CLAIM-3DC9A330ACED56ED
semantic_clauses:
  - Each market index and date must keep the level delivered last.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 54f7504574e867e6acd579fd94445ae641eae26861e5f4fab20a3bd2d10ad1e8
semantic_inventory:
  - claim_key: CLAIM-3DC9A330ACED56ED
    claim_text: Each market index and date must keep the level delivered last
    role: normative
    status: modeled
    span:
      start: 0
      end: 61
id: REQ-valuation-index-level-last-delivered
type: req
---
Each market index and date must keep the level delivered last.

## Context

Market index levels aren't ranked (REQ-valuation-index-levels-not-ranked), so when two deliveries carry a level for the same index and date, something must decide which is kept. The implementation raised this as question #8 on docs/implementation-questions.md and stores one level per index and date, the last delivered replacing an earlier one (REQ-valuation-index-levels, which the author has not reviewed). On 2026-10-31 the author confirmed that as D-124.

## Source

> A source in D-94's ranking is one connector instance, ranked by default by the institution it reports, with the first-party connector of each kind first. Cross-check prices never rank: they only check (D-22). Market index levels aren't ranked; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#5
