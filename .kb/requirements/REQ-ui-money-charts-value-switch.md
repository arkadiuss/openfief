---
title: The allocation and value-over-time charts must each have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-6
  - lane:strict
text_ref: design-decisions:D-77#6
origin:
  kind: import
  ref: design-decisions:D-77#6
  recorded_at: '2026-10-29T10:32:06.783Z'
semantic_text: The allocation and value-over-time charts must each have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
logic_claims:
  - CLAIM-BE6A574E242CD039
semantic_clauses:
  - The allocation and value-over-time charts must each have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 52a651f29cf3fc4351059d973945e77edaf9237610437c947c8058a62973dc56
semantic_inventory:
  - claim_key: CLAIM-BE6A574E242CD039
    claim_text: The allocation and value-over-time charts must each have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines
    role: normative
    status: modeled
    span:
      start: 0
      end: 200
id: REQ-ui-money-charts-value-switch
type: req
---
The allocation and value-over-time charts must each have a switch between the gross, net and liquidation values that starts on gross, and the value-over-time chart must be able to draw all three lines.

## Why this changed

REQ-ui-value-switch said every chart has the switch, following D-77's first wording. In the same session the author decided D-106: the performance chart is a time-weighted return index computed on gross value, so it has no switch. D-77 was narrowed to the money charts, allocation and value over time, to agree with D-106, and this requirement states the narrowed rule.

## Context

D-77 gives every holding, account and total three values (REQ-valuation-three-values), and tables show all three side by side (REQ-ui-three-values-tables). D-77 replaces D-25, under which bonds counted in totals at their net-if-redeemed-today value by default (REQ-AUTO-E10BE8CE13E67657, superseded through REQ-ui-value-switch).

## Source

> Tables show all three side by side. The allocation and value-over-time charts each have a switch between them that starts on gross, the value broker statements show, and the value-over-time chart (D-30) can draw all three lines. The performance chart is computed on gross value (D-106).

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-77#6
