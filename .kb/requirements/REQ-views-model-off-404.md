---
title: The read API must answer 404 for the model portfolio while it is off.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-8
  - lane:strict
text_ref: design-decisions:D-107#8
origin:
  kind: import
  ref: design-decisions:D-107#8
  recorded_at: '2026-10-29T10:20:46.676Z'
semantic_text: The read API must answer 404 for the model portfolio while it is off.
logic_claims:
  - CLAIM-857745BF11A3EA74
semantic_clauses:
  - The read API must answer 404 for the model portfolio while it is off.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d408c51f692a76ee9e3a7d17ab32b46c4ad1bc5df783a59c2da5e297f8f267ad
semantic_inventory:
  - claim_key: CLAIM-857745BF11A3EA74
    claim_text: The read API must answer 404 for the model portfolio while it is off
    role: normative
    status: modeled
    span:
      start: 0
      end: 68
id: REQ-views-model-off-404
type: req
---
The read API must answer 404 for the model portfolio while it is off.

## Context

D-52 keeps the model portfolio out of the UI and the read API while it is off (REQ-portfolio-model-optional). D-107 records, as the author's decision, what the implementation already answers then: 404. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#8
