---
title: The built-in UI must show an allocation chart by asset class, currency, broker or sleeve.
status: open
text_ref: design-decisions:D-30#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-30-1
  - lane:strict
semantic_text: The built-in UI must show an allocation chart by asset class, currency, broker or sleeve.
logic_claims:
  - CLAIM-355F92E4974BBDC5
semantic_clauses:
  - The built-in UI must show an allocation chart by asset class, currency, broker or sleeve.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a7e46fee40e1b935405a1db439ca59825ecc43ff4107503da209d4f80bf54e9
semantic_inventory:
  - claim_key: CLAIM-355F92E4974BBDC5
    claim_text: The built-in UI must show an allocation chart by asset class, currency, broker or sleeve
    role: normative
    status: modeled
    span:
      start: 0
      end: 88
origin:
  kind: import
  ref: design-decisions:D-30#1
  recorded_at: '2026-10-27T20:12:50.546Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-68636F42D98D4CF7
type: req
---
The built-in UI must show an allocation chart by asset class, currency, broker or sleeve.

## Context

D-30 names the charts the built-in UI draws in M0. It first named two, allocation and value over time; with D-106 the author separated value and performance, and D-30 now names three. The allocation chart is unchanged.

## Source

> The built-in UI draws three charts in M0: allocation (by asset class, currency, broker or sleeve), value over time, and performance (D-106). The time charts start at go-live (D-23).

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-30#1
