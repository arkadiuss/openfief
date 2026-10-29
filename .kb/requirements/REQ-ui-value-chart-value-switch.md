---
title: The value-over-time chart must have D-77's switch between gross, net and liquidation value.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-106-3
  - lane:strict
text_ref: design-decisions:D-106#3
origin:
  kind: import
  ref: design-decisions:D-106#3
  recorded_at: '2026-10-29T10:13:42.392Z'
semantic_text: The value-over-time chart must have D-77's switch between gross, net and liquidation value.
logic_claims:
  - CLAIM-233374B15710AA19
semantic_clauses:
  - The value-over-time chart must have D-77's switch between gross, net and liquidation value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 50671b5fc8c85f11e654dbf102b8a433deb90c344bbff7e295f9d2ab0a0f1fbf
semantic_inventory:
  - claim_key: CLAIM-233374B15710AA19
    claim_text: The value-over-time chart must have D-77's switch between gross, net and liquidation value
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
id: REQ-ui-value-chart-value-switch
type: req
---
The value-over-time chart must have D-77's switch between gross, net and liquidation value.

## Context

D-77 gives every holding, account and total three values and a switch between them on charts (REQ-ui-value-switch). D-106 keeps that switch on the value-over-time chart, while the performance chart is computed on gross value (REQ-ui-performance-chart-twr-index). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106#3
