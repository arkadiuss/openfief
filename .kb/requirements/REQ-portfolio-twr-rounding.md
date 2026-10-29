---
title: The time-weighted return must round each daily factor and the running product half-even to 12 decimal places.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-2
  - lane:strict
text_ref: design-decisions:D-107#2
origin:
  kind: import
  ref: design-decisions:D-107#2
  recorded_at: '2026-10-29T10:16:43.895Z'
semantic_text: The time-weighted return must round each daily factor and the running product half-even to 12 decimal places.
logic_claims:
  - CLAIM-7F347E90C3A4F49F
semantic_clauses:
  - The time-weighted return must round each daily factor and the running product half-even to 12 decimal places.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8b9c5d34a981d46f4650659a2cabf6929233cd22d310d25009e28726c87be544
semantic_inventory:
  - claim_key: CLAIM-7F347E90C3A4F49F
    claim_text: The time-weighted return must round each daily factor and the running product half-even to 12 decimal places
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
id: REQ-portfolio-twr-rounding
type: req
---
The time-weighted return must round each daily factor and the running product half-even to 12 decimal places.

## Context

D-61 measures the portfolio's line as a time-weighted return. D-107 records, as the author's decisions, conventions the implementation had already built for it and for the model portfolio, including this rounding. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#2
