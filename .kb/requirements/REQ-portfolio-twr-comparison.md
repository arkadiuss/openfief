---
title: Next to the model, the portfolio's line must be its time-weighted return, so deposits and withdrawals don't count as performance.
status: open
text_ref: design-decisions:D-61#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-61-1
  - lane:strict
semantic_text: Next to the model, the portfolio's line must be its time-weighted return, so deposits and withdrawals don't count as performance.
logic_claims:
  - CLAIM-6442521669FF1B23
semantic_clauses:
  - Next to the model, the portfolio's line must be its time-weighted return, so deposits and withdrawals don't count as performance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a9bbd6cc91f13bb2cfd475970767472f1beafea91a1e425b4dbd1ec09e58be37
semantic_inventory:
  - claim_key: CLAIM-6442521669FF1B23
    claim_text: Next to the model, the portfolio's line must be its time-weighted return, so deposits and withdrawals don't count as performance
    role: normative
    status: modeled
    span:
      start: 0
      end: 128
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-61#1
  recorded_at: '2026-10-28T06:18:38.589Z'
id: REQ-portfolio-twr-comparison
type: req
---
Next to the model, the portfolio's line must be its time-weighted return, so deposits and withdrawals don't count as performance.

## Context

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

## Source

> Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows. Both lines are in the reporting currency (D-62). The model's backtest also needs exchange rates into that currency, so it starts no earlier than they do; NBP's rates begin on 2002-01-02.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-61#1
