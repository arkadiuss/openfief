---
title: While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates.
status: open
text_ref: design-decisions:D-200#19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-19
  - lane:strict
semantic_text: While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates.
logic_claims:
  - CLAIM-F903442083B68711
semantic_clauses:
  - While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 370a40c0c556bcadb9bf297c8a3b90a2a522bdba9f9052ed37539894c31d235d
semantic_inventory:
  - claim_key: CLAIM-F903442083B68711
    claim_text: While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates
    role: normative
    status: modeled
    span:
      start: 0
      end: 295
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#19
  recorded_at: '2026-11-07T03:32:55.039Z'
id: REQ-m0-exit-19-model-portfolio
type: req
---
While the model portfolio is off, the UI and the read API must show nothing of it; when it is on, the backtest of a two-target basket over a synthetic price series must match a hand calculation across one yearly rebalance, and each max drawdown must match the series' known peak and trough dates.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-19.

## Source

> **Model portfolio.** While off, the UI and the read API show nothing of it. When on, the backtest of a two-target basket over a synthetic price series matches a hand calculation across one yearly rebalance, and each max drawdown matches the series' known peak and trough dates.

design-decisions:D-200#19 (docs/DESIGN.md, D-200, criterion 19)
