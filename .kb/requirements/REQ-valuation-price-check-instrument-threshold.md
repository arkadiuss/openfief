---
title: The owner must be able to set a different price-check threshold for a single instrument.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-110-2
  - lane:strict
text_ref: design-decisions:D-110#2
origin:
  kind: import
  ref: design-decisions:D-110#2
  recorded_at: '2026-10-29T10:52:25.737Z'
semantic_text: The owner must be able to set a different price-check threshold for a single instrument.
logic_claims:
  - CLAIM-FA81D47CE64D7E69
semantic_clauses:
  - The owner must be able to set a different price-check threshold for a single instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b57891e9cd243cae9f8488ca8fe1fb28a912e5e198b28a9f7cb6e00393ebecc3
semantic_inventory:
  - claim_key: CLAIM-FA81D47CE64D7E69
    claim_text: The owner must be able to set a different price-check threshold for a single instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
id: REQ-valuation-price-check-instrument-threshold
type: req
---
The owner must be able to set a different price-check threshold for a single instrument.

## Why this changed

The author's decision D-110 sets the flag threshold at 1% and lets the owner set a different threshold for a single instrument, such as a thinly traded Warsaw listing. REQ-valuation-price-check-threshold made the threshold one setting of the core for every instrument, so this requirement replaces it; the 1% rule itself is REQ-valuation-price-check-one-percent.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data prices. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A cross-check price is flagged when it differs from the stored price by more than 1% of it, in either direction. The owner can set a different threshold for a single instrument, such as a thinly traded Warsaw listing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-110#2
