---
title: Model portfolio weights must go in steps of 0.01%.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-6
  - lane:strict
text_ref: design-decisions:D-107#6
origin:
  kind: import
  ref: design-decisions:D-107#6
  recorded_at: '2026-10-29T10:19:30.900Z'
semantic_text: Model portfolio weights must go in steps of 0.01%.
logic_claims:
  - CLAIM-04FB01433C2FF03B
semantic_clauses:
  - Model portfolio weights must go in steps of 0.01%.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 40c8a2b9b091cb5894a03f2d1a684af64df4ad9b361df5e0550c298a0f9500b1
semantic_inventory:
  - claim_key: CLAIM-04FB01433C2FF03B
    claim_text: Model portfolio weights must go in steps of 0.01%
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-portfolio-model-weight-step
type: req
---
Model portfolio weights must go in steps of 0.01%.

## Context

D-52 makes the model portfolio a weighted basket whose weights sum to 100% (REQ-portfolio-model-weights). D-107 records, as the author's decision, the weight step the implementation had already built. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#6
