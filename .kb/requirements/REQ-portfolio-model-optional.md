---
title: The model portfolio must be off by default, and while it is off it must appear nowhere in the UI or the read API.
status: open
text_ref: design-decisions:D-52#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-52-1
  - lane:strict
semantic_text: The model portfolio must be off by default, and while it is off it must appear nowhere in the UI or the read API.
logic_claims:
  - CLAIM-9A6F34A63EA9E2B3
semantic_clauses:
  - The model portfolio must be off by default, and while it is off it must appear nowhere in the UI or the read API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 477d81a017206ef239ee51522dfdf23bef7ddde72c391d55926dd1bfb9b82e1f
semantic_inventory:
  - claim_key: CLAIM-9A6F34A63EA9E2B3
    claim_text: The model portfolio must be off by default, and while it is off it must appear nowhere in the UI or the read API
    role: normative
    status: modeled
    span:
      start: 0
      end: 112
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-52#1
  recorded_at: '2026-10-28T06:08:59.276Z'
id: REQ-portfolio-model-optional
type: req
---
The model portfolio must be off by default, and while it is off it must appear nowhere in the UI or the read API.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-52#1
