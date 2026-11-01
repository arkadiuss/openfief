---
title: 'An acknowledgement must cover one flagged price check: those two prices on their dates, so a changed or newer price is checked anew.'
status: open
text_ref: design-decisions:D-142#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-142-1
  - lane:strict
semantic_text: 'An acknowledgement must cover one flagged price check: those two prices on their dates, so a changed or newer price is checked anew.'
logic_claims:
  - CLAIM-B9C5971E67CE7BC9
semantic_clauses:
  - 'An acknowledgement must cover one flagged price check: those two prices on their dates, so a changed or newer price is checked anew.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 733f593f53f5a84241bac15b0c1a0b3d6348c05f3501c9f3a4755f2f1806f89b
semantic_inventory:
  - claim_key: CLAIM-B9C5971E67CE7BC9
    claim_text: 'An acknowledgement must cover one flagged price check: those two prices on their dates, so a changed or newer price is checked anew'
    role: normative
    status: modeled
    span:
      start: 0
      end: 131
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-142#1
  recorded_at: '2026-11-01T12:24:55.335Z'
id: REQ-valuation-acknowledgement-covers-one-check
type: req
---
An acknowledgement must cover one flagged price check: those two prices on their dates, so a changed or newer price is checked anew.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #26. D-112 lets the owner acknowledge a flagged check, recorded append-only (REQ-valuation-price-check-acknowledge); this says how far one acknowledgement reaches. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An acknowledgement covers one flagged price check: those two prices on their dates, so a changed or newer price is checked anew. An instrument's threshold is a fraction greater than 0 and less than 1, and clearing it restores 1%.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-142#1
