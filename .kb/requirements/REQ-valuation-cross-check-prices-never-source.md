---
title: 'Cross-check prices, such as IBKR''s mark prices, must never be a source of a series'' values: they only check.'
status: open
text_ref: design-decisions:D-124#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-124-3
  - lane:strict
semantic_text: 'Cross-check prices, such as IBKR''s mark prices, must never be a source of a series'' values: they only check.'
logic_claims:
  - CLAIM-AC14827E6ECB6CEF
semantic_clauses:
  - 'Cross-check prices, such as IBKR''s mark prices, must never be a source of a series'' values: they only check.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 927b406ddfcea49f7eae8773227763060c6393cb5fe3fe89a0fcc5f6d32b5749
semantic_inventory:
  - claim_key: CLAIM-AC14827E6ECB6CEF
    claim_text: 'Cross-check prices, such as IBKR''s mark prices, must never be a source of a series'' values: they only check'
    role: normative
    status: modeled
    span:
      start: 0
      end: 107
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-124#3
  recorded_at: '2026-11-01T07:37:37.770Z'
id: REQ-valuation-cross-check-prices-never-source
type: req
---
Cross-check prices, such as IBKR's mark prices, must never be a source of a series' values: they only check.

## Why this changed

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. REQ-valuation-cross-check-prices-never-rank kept cross-check prices out of D-94's per-date ranking; with one source per series, they are never a series' source at all, so this requirement replaces it.

## Context

D-22 has IBKR Flex prices serve as a cross-check, and D-112 lets only a connector whose manifest declares it send cross-check prices. A source is one connector instance (REQ-valuation-market-data-source-instance). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A source in D-94 is one connector instance. Cross-check prices, such as IBKR's mark prices, are never a source: they only check (D-22). Market index levels have no ranking; each index and date keeps the value delivered last.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-124#3
