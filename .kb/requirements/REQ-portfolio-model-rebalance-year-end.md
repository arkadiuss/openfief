---
title: The model portfolio must rebalance at the end of each 31 December.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-107-4
  - lane:strict
text_ref: design-decisions:D-107#4
origin:
  kind: import
  ref: design-decisions:D-107#4
  recorded_at: '2026-10-29T10:18:01.710Z'
semantic_text: The model portfolio must rebalance at the end of each 31 December.
logic_claims:
  - CLAIM-18868D1AA5549AE7
semantic_clauses:
  - The model portfolio must rebalance at the end of each 31 December.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 688dbbfc4262a1a78fa7017cf3b0b53c6572ae732d65dea292e9c17095a269af
semantic_inventory:
  - claim_key: CLAIM-18868D1AA5549AE7
    claim_text: The model portfolio must rebalance at the end of each 31 December
    role: normative
    status: modeled
    span:
      start: 0
      end: 65
id: REQ-portfolio-model-rebalance-year-end
type: req
---
The model portfolio must rebalance at the end of each 31 December.

## Context

D-53 rebalances the model portfolio's backtest to the target weights once a year (REQ-portfolio-backtest-yearly-rebalance). D-107 records, as the author's decision, the convention the implementation had already built: the rebalance happens at the end of each 31 December. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Conventions as built: in the time-weighted return, cash flows count at the end of their day, each daily factor and the running product are rounded half-even to 12 decimal places, and opening lots count as inflows at their market value on their date. The model portfolio rebalances at the end of each 31 December. A target's drawdown is measured in the reporting currency over the span where it has both closes and exchange rates. Weights go in steps of 0.01%, with at most 50 targets, and while the model is off the read API answers 404 for it.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-107#4
