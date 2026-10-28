---
title: The first market-data connector must use EODHD.
status: open
text_ref: design-decisions:D-51#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-51-1
  - lane:strict
semantic_text: The first market-data connector must use EODHD.
logic_claims:
  - CLAIM-A9CDFEB1E1C2D30F
semantic_clauses:
  - The first market-data connector must use EODHD.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5f2d8fbdb8b28227287d2c63d79c9243e33b00cb334900c46a5d424e7475f2af
semantic_inventory:
  - claim_key: CLAIM-A9CDFEB1E1C2D30F
    claim_text: The first market-data connector must use EODHD
    role: normative
    status: modeled
    span:
      start: 0
      end: 46
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-51#1
  recorded_at: '2026-10-28T06:08:55.075Z'
id: REQ-valuation-eodhd-connector
type: req
---
The first market-data connector must use EODHD.

## Context

XTB shut down its public API in 2025, and neither the treasury bond registry nor the bank that sells the bonds offers one, so both arrive as files or manual entry. Interactive Brokers' Flex Queries are expected to report positions per lot, with open date and cost (OQ-22 verifies this).

## Source

> The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-51#1
