---
title: The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year.
status: open
text_ref: design-decisions:D-53#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-53-1
  - lane:strict
semantic_text: The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year.
logic_claims:
  - CLAIM-3D89D9DD3B0B389D
semantic_clauses:
  - The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0cae5351c32c42f077970518d3b886b552771524e35a7e300ada656a18d3c4fc
semantic_inventory:
  - claim_key: CLAIM-3D89D9DD3B0B389D
    claim_text: The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year
    role: normative
    status: modeled
    span:
      start: 0
      end: 153
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-53#1
  recorded_at: '2026-10-28T06:09:06.945Z'
id: REQ-portfolio-backtest-yearly-rebalance
type: req
---
The model portfolio's backtest must be computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100. How the portfolio's own line is measured is open (OQ-28).

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-53#1
