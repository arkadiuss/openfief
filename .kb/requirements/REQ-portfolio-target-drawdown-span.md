---
title: A target's drawdown must be measured in the reporting currency over the span where it has both closes and exchange rates.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-5
  - lane:strict
text_ref: design-decisions:D-107#5
origin:
  kind: import
  ref: design-decisions:D-107#5
  recorded_at: '2026-10-29T10:18:52.527Z'
semantic_text: A target's drawdown must be measured in the reporting currency over the span where it has both closes and exchange rates.
logic_claims:
  - CLAIM-A2EFEF43F98507EB
semantic_clauses:
  - A target's drawdown must be measured in the reporting currency over the span where it has both closes and exchange rates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 18a6f27acc95bbea68ffe8868981d169bd92a188e4a85fe7f47764e71efd23cc
semantic_inventory:
  - claim_key: CLAIM-A2EFEF43F98507EB
    claim_text: A target's drawdown must be measured in the reporting currency over the span where it has both closes and exchange rates
    role: normative
    status: modeled
    span:
      start: 0
      end: 120
id: REQ-portfolio-target-drawdown-span
type: req
---
A target's drawdown must be measured in the reporting currency over the span where it has both closes and exchange rates.

## Context

D-54 shows the historical max drawdown of the model and of each target (REQ-portfolio-max-drawdown), and D-62 shows the model comparison in the reporting currency. D-107 records, as the author's decision, how the implementation already measures a target's drawdown. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#5
