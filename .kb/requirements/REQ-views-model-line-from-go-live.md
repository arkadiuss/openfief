---
title: When the model portfolio is on, the value-over-time chart must show it as a second line from go-live, with both lines starting at 100.
status: open
text_ref: design-decisions:D-53#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-53-2
  - lane:strict
semantic_text: When the model portfolio is on, the value-over-time chart must show it as a second line from go-live, with both lines starting at 100.
logic_claims:
  - CLAIM-4CD7474197A4689C
semantic_clauses:
  - When the model portfolio is on, the value-over-time chart must show it as a second line from go-live, with both lines starting at 100.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 218f7c3a44a0feeaac1e2a87b3b73c351739b2bfefdda98a44887a5d0d148631
semantic_inventory:
  - claim_key: CLAIM-4CD7474197A4689C
    claim_text: When the model portfolio is on, the value-over-time chart must show it as a second line from go-live, with both lines starting at 100
    role: condition
    status: modeled
    span:
      start: 0
      end: 133
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-53#2
  recorded_at: '2026-10-28T06:09:11.105Z'
id: REQ-views-model-line-from-go-live
type: req
---
When the model portfolio is on, the value-over-time chart must show it as a second line from go-live, with both lines starting at 100.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100. How the portfolio's own line is measured is open (OQ-28).

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-53#2
