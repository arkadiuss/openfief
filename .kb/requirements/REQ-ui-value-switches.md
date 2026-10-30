---
title: The overview's allocation and value-over-time chart must each start on gross with a switch to net and liquidation, the value chart also offering all three lines, and the value chart must draw the portfolio's daily totals with cash from go-live in the reporting currency, skipping days without the value picked.
status: open
text_ref: docs/DESIGN.md#D-77
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-77
  - lane:strict
semantic_text: The overview's allocation and value-over-time chart must each start on gross with a switch to net and liquidation, the value chart also offering all three lines, and the value chart must draw the portfolio's daily totals with cash from go-live in the reporting currency, skipping days without the value picked.
logic_claims:
  - CLAIM-BD6DA8FF56DC9F4F
semantic_clauses:
  - The overview's allocation and value-over-time chart must each start on gross with a switch to net and liquidation, the value chart also offering all three lines, and the value chart must draw the portfolio's daily totals with cash from go-live in the reporting currency, skipping days without the value picked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3b5c8d2ed3d48f7fdadb1644a5368aa992bfe3a37a53e9efb770bd3cace22f85
semantic_inventory:
  - claim_key: CLAIM-BD6DA8FF56DC9F4F
    claim_text: The overview's allocation and value-over-time chart must each start on gross with a switch to net and liquidation, the value chart also offering all three lines, and the value chart must draw the portfolio's daily totals with cash from go-live in the reporting currency, skipping days without the value picked
    role: normative
    status: modeled
    span:
      start: 0
      end: 309
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:34.939Z'
id: REQ-ui-value-switches
type: req
---
The overview's allocation and value-over-time chart must each start on gross with a switch to net and liquidation, the value chart also offering all three lines, and the value chart must draw the portfolio's daily totals with cash from go-live in the reporting currency, skipping days without the value picked.

## Context

D-77 gives both money charts a switch that starts on gross, and the value chart all three lines; D-106 draws money with cash. Skipping a day without the value rather than guessing it follows D-78. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-106]** Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

docs/DESIGN.md#D-106
