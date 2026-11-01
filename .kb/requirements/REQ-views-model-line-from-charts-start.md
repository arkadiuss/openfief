---
title: The performance chart must show the model portfolio as a second line from the charts' start, with both lines starting at 100.
status: open
text_ref: design-decisions:D-53#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-53-2
  - lane:strict
semantic_text: The performance chart must show the model portfolio as a second line from the charts' start, with both lines starting at 100.
logic_claims:
  - CLAIM-A4FD0E1171D9361B
semantic_clauses:
  - The performance chart must show the model portfolio as a second line from the charts' start, with both lines starting at 100.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 158e1b467a5112dee5cd5f9b3ce256fd373c26d22feee6c2511ab5b6f0a48eca
semantic_inventory:
  - claim_key: CLAIM-A4FD0E1171D9361B
    claim_text: The performance chart must show the model portfolio as a second line from the charts' start, with both lines starting at 100
    role: normative
    status: modeled
    span:
      start: 0
      end: 124
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-53#2
  recorded_at: '2026-11-01T07:38:28.734Z'
id: REQ-views-model-line-from-charts-start
type: req
---
The performance chart must show the model portfolio as a second line from the charts' start, with both lines starting at 100.

## Why this changed

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. REQ-views-model-line-on-performance-chart drew the model's line from go-live, so this requirement replaces it.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand, and D-52 keeps it out of the UI while it is off (REQ-portfolio-model-optional). Its line is a backtest (REQ-portfolio-backtest-yearly-rebalance). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The performance chart (D-106) shows it as a second line from the charts' start (D-138), with both lines starting at 100. The portfolio's line is measured as D-61 says.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-53#2
