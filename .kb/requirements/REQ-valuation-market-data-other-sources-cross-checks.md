---
title: The values of every source other than a series' own source must be stored as cross-checks.
status: open
text_ref: design-decisions:D-94#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-3
  - lane:strict
semantic_text: The values of every source other than a series' own source must be stored as cross-checks.
logic_claims:
  - CLAIM-97AB68D9F08A809C
semantic_clauses:
  - The values of every source other than a series' own source must be stored as cross-checks.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 34d825516d66971b35aac3765784b6ea1bdef7c86f8165adc79e8bffec7f8c5f
semantic_inventory:
  - claim_key: CLAIM-97AB68D9F08A809C
    claim_text: The values of every source other than a series' own source must be stored as cross-checks
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#3
  recorded_at: '2026-11-01T07:36:54.915Z'
id: REQ-valuation-market-data-other-sources-cross-checks
type: req
---
The values of every source other than a series' own source must be stored as cross-checks.

## Why this changed

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. REQ-valuation-market-data-all-sources-stored kept every source's values so that the lower-ranked ones served as cross-checks next to a per-date ranking; with one source per series, the other sources' values are stored only to check it, so this requirement replaces it.

## Context

D-22 has IBKR's prices serve as a cross-check of the market-data connector's, and D-110 to D-112 say how a cross-check is compared and flagged. A stored cross-check value never becomes the series' value (D-94). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#3
