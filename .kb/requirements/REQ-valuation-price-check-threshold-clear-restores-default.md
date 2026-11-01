---
title: Clearing an instrument's price-check threshold must restore 1%.
status: open
text_ref: design-decisions:D-142#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-142-3
  - lane:strict
semantic_text: Clearing an instrument's price-check threshold must restore 1%.
logic_claims:
  - CLAIM-92D5F0B3852C7332
semantic_clauses:
  - Clearing an instrument's price-check threshold must restore 1%.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1091767b2203ca59d4d88849b7c5c3270b530c0279c4b648987a493c68c70c76
semantic_inventory:
  - claim_key: CLAIM-92D5F0B3852C7332
    claim_text: Clearing an instrument's price-check threshold must restore 1%
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-142#3
  recorded_at: '2026-11-01T12:25:13.003Z'
id: REQ-valuation-price-check-threshold-clear-restores-default
type: req
---
Clearing an instrument's price-check threshold must restore 1%.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #26. 1% is D-110's default threshold (REQ-valuation-price-check-one-percent). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An acknowledgement covers one flagged price check: those two prices on their dates, so a changed or newer price is checked anew. An instrument's threshold is a fraction greater than 0 and less than 1, and clearing it restores 1%.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-142#3
