---
title: The built-in UI must show a performance chart that starts at go-live.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-30-3
  - lane:strict
text_ref: design-decisions:D-30#3
origin:
  kind: import
  ref: design-decisions:D-30#3
  recorded_at: '2026-10-29T10:07:40.142Z'
semantic_text: The built-in UI must show a performance chart that starts at go-live.
logic_claims:
  - CLAIM-5BE430A731053DC8
semantic_clauses:
  - The built-in UI must show a performance chart that starts at go-live.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2eb3973edc1eee01172af3e20b2f4c7028c7a8712f2f3d0a5bd06f992c6caca2
semantic_inventory:
  - claim_key: CLAIM-5BE430A731053DC8
    claim_text: The built-in UI must show a performance chart that starts at go-live
    role: normative
    status: modeled
    span:
      start: 0
      end: 68
id: REQ-ui-performance-chart
type: req
---
The built-in UI must show a performance chart that starts at go-live.

## Context

D-30 named two charts, allocation and value over time. With D-106 the author made value and performance separate charts, and amended D-30 to name three: allocation, value over time and performance, with both time charts starting at go-live (D-23). REQ-AUTO-68636F42D98D4CF7 and REQ-AUTO-718AA30AAFD0E6E8 hold the other two charts, and D-106's requirements say what the performance chart shows. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> The built-in UI draws three charts in M0: allocation (by asset class, currency, broker or sleeve), value over time, and performance (D-106). The time charts start at go-live (D-23).

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-30#3
