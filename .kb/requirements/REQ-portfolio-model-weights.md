---
title: The model portfolio's target weights must sum to 100%.
status: open
text_ref: design-decisions:D-52#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-52-2
  - lane:strict
semantic_text: The model portfolio's target weights must sum to 100%.
logic_claims:
  - CLAIM-F0815251D00D7294
semantic_clauses:
  - The model portfolio's target weights must sum to 100%.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e7f5f9ab663e92ffd768cfe6a3b46cfc9f6c597e39b4258d14e4241e9a3bc1b2
semantic_inventory:
  - claim_key: CLAIM-F0815251D00D7294
    claim_text: The model portfolio's target weights must sum to 100%
    role: normative
    status: modeled
    span:
      start: 0
      end: 53
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-52#2
  recorded_at: '2026-10-28T06:09:02.443Z'
id: REQ-portfolio-model-weights
type: req
---
The model portfolio's target weights must sum to 100%.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-52#2
