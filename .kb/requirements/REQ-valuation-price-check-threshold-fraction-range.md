---
title: An instrument's price-check threshold must be a fraction greater than 0 and less than 1.
status: open
text_ref: design-decisions:D-142#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-142-2
  - lane:strict
semantic_text: An instrument's price-check threshold must be a fraction greater than 0 and less than 1.
logic_claims:
  - CLAIM-789E7F72253BE921
semantic_clauses:
  - An instrument's price-check threshold must be a fraction greater than 0 and less than 1.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b8f6540f5af4bcb0fdd84142e69815be77a4f59fe5c0f2ae3ad5b46226f20207
semantic_inventory:
  - claim_key: CLAIM-789E7F72253BE921
    claim_text: An instrument's price-check threshold must be a fraction greater than 0 and less than 1
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-142#2
  recorded_at: '2026-11-01T12:25:04.298Z'
id: REQ-valuation-price-check-threshold-fraction-range
type: req
---
An instrument's price-check threshold must be a fraction greater than 0 and less than 1.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #26. D-110 flags a cross-check price beyond 1% of the stored price and lets the owner set a different threshold for one instrument (REQ-valuation-price-check-instrument-threshold). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An acknowledgement covers one flagged price check: those two prices on their dates, so a changed or newer price is checked anew. An instrument's threshold is a fraction greater than 0 and less than 1, and clearing it restores 1%.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-142#2
