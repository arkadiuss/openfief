---
title: The time-weighted return must count opening lots as inflows at their market value on their date.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-3
  - lane:strict
text_ref: design-decisions:D-107#3
origin:
  kind: import
  ref: design-decisions:D-107#3
  recorded_at: '2026-10-29T10:17:22.904Z'
semantic_text: The time-weighted return must count opening lots as inflows at their market value on their date.
logic_claims:
  - CLAIM-07EC08F71F778541
semantic_clauses:
  - The time-weighted return must count opening lots as inflows at their market value on their date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 06c3d842aabb9c30398c11d77398826483330252e2fe583f027dda260f3b9955
semantic_inventory:
  - claim_key: CLAIM-07EC08F71F778541
    claim_text: The time-weighted return must count opening lots as inflows at their market value on their date
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
id: REQ-portfolio-twr-opening-lots-inflows
type: req
---
The time-weighted return must count opening lots as inflows at their market value on their date.

## Context

D-23 seeds accounts with opening lots, and D-98 lets a manual account receive them one at a time. D-107 records, as the author's decisions, conventions the implementation had already built for the time-weighted return of D-61, including how those lots enter it. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#3
