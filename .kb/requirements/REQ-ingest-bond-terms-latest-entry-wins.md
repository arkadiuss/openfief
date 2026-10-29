---
title: The latest series terms the owner enters with a lot must win for that series.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-3
  - lane:strict
text_ref: design-decisions:D-97#3
origin:
  kind: import
  ref: design-decisions:D-97#3
  recorded_at: '2026-10-29T09:46:27.902Z'
semantic_text: The latest series terms the owner enters with a lot must win for that series.
logic_claims:
  - CLAIM-776833D4252C4BA8
semantic_clauses:
  - The latest series terms the owner enters with a lot must win for that series.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 29445ad61bee6d33a7ddfe25f92e35c3d5acd20a0c825443c0d069f2c9924075
semantic_inventory:
  - claim_key: CLAIM-776833D4252C4BA8
    claim_text: The latest series terms the owner enters with a lot must win for that series
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
id: REQ-ingest-bond-terms-latest-entry-wins
type: req
---
The latest series terms the owner enters with a lot must win for that series.

## Context

Until OQ-30 is answered, D-97 has the owner enter a series' terms with each lot (REQ-ingest-bond-terms-entered-with-lot), so the same series can be entered several times. The author settles which entry counts: the latest one for the series. This is about bond series terms the owner types in, not market data, where D-94 ranks sources instead. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The treasury bond connector reads each series' terms (first-period rate, margin and early-redemption fee) from the issuer's site, obligacjeskarbowe.pl, once its terms of use are confirmed to allow it (OQ-30, D-8). Until then the owner enters them with a lot, and the latest entry for a series wins.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#3
