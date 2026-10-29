---
title: The model portfolio must have at most 50 targets.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-7
  - lane:strict
text_ref: design-decisions:D-107#7
origin:
  kind: import
  ref: design-decisions:D-107#7
  recorded_at: '2026-10-29T10:20:06.722Z'
semantic_text: The model portfolio must have at most 50 targets.
logic_claims:
  - CLAIM-87020BCAAE1F3508
semantic_clauses:
  - The model portfolio must have at most 50 targets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 529539061a0ea7913ab9989e2fc6cb3125b665547cad2110b97134619697773f
semantic_inventory:
  - claim_key: CLAIM-87020BCAAE1F3508
    claim_text: The model portfolio must have at most 50 targets
    role: normative
    status: modeled
    span:
      start: 0
      end: 48
id: REQ-portfolio-model-max-targets
type: req
---
The model portfolio must have at most 50 targets.

## Context

D-52 makes the model portfolio a weighted basket of targets. D-107 records, as the author's decision, the limit on targets the implementation had already built. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#7
