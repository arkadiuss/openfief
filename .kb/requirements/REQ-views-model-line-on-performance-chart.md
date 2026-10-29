---
title: The performance chart must show the model portfolio as a second line from go-live, with both lines starting at 100.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-53-2
  - lane:strict
text_ref: design-decisions:D-53#2
origin:
  kind: import
  ref: design-decisions:D-53#2
  recorded_at: '2026-10-29T10:08:32.499Z'
semantic_text: The performance chart must show the model portfolio as a second line from go-live, with both lines starting at 100.
logic_claims:
  - CLAIM-14A2C3A7FF2153AC
semantic_clauses:
  - The performance chart must show the model portfolio as a second line from go-live, with both lines starting at 100.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8a48ff960d5288a44bb2d611a177c665d0b2da1be9d1fe334b2d6398e6e77eb3
semantic_inventory:
  - claim_key: CLAIM-14A2C3A7FF2153AC
    claim_text: The performance chart must show the model portfolio as a second line from go-live, with both lines starting at 100
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
id: REQ-views-model-line-on-performance-chart
type: req
---
The performance chart must show the model portfolio as a second line from go-live, with both lines starting at 100.

## Why this changed

The author's decision D-106 makes value and performance separate charts, and D-53 was amended with it so that the model portfolio's line is drawn on the performance chart (D-106) rather than on the value-over-time chart. REQ-views-model-line-from-go-live put the model's line on the value-over-time chart, so this requirement replaces it; the line still starts at go-live, with both lines starting at 100.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket. D-52 keeps the model out of the UI while it is off (REQ-portfolio-model-optional), and D-106 has the performance chart show the portfolio's time-weighted return next to it. The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The performance chart (D-106) shows it as a second line from go-live, with both lines starting at 100. The portfolio's line is measured as D-61 says.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-53#2

> Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-106
