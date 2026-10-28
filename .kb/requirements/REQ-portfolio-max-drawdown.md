---
title: M0 must show the historical max drawdown of the model portfolio and of each target, with its peak and trough dates.
status: open
text_ref: design-decisions:D-54#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-54-1
  - lane:strict
semantic_text: M0 must show the historical max drawdown of the model portfolio and of each target, with its peak and trough dates.
logic_claims:
  - CLAIM-5506F57E1D9F7417
semantic_clauses:
  - M0 must show the historical max drawdown of the model portfolio and of each target, with its peak and trough dates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6de0b3ec0f1b3eaa695e0a05f8cb908d2e72a491b4a81553b14251836b24274c
semantic_inventory:
  - claim_key: CLAIM-5506F57E1D9F7417
    claim_text: M0 must show the historical max drawdown of the model portfolio and of each target, with its peak and trough dates
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-54#1
  recorded_at: '2026-10-28T06:09:15.573Z'
id: REQ-portfolio-max-drawdown
type: req
---
M0 must show the historical max drawdown of the model portfolio and of each target, with its peak and trough dates.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> M0 shows the historical max drawdown of the model and of each target: the worst fall from a peak to a later trough over its whole history, with both dates. M0 has no stress tests and no statistical risk estimates.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-54#1
