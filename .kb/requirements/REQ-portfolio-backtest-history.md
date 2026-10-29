---
title: Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none.
status: open
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
semantic_text: Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none.
logic_claims:
  - CLAIM-A33ED550CC477AF2
semantic_clauses:
  - Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2e0335d546ab631a4c69ae05d745453c3d8db62c982608e8681f7e16eeae2b15
semantic_inventory:
  - claim_key: CLAIM-A33ED550CC477AF2
    claim_text: Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none
    role: normative
    status: modeled
    span:
      start: 0
      end: 293
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:40.170Z'
id: REQ-portfolio-backtest-history
type: req
---
Each model target's history must be its latest end-of-day close or level times the latest FX rate into the reporting currency for each day from the first day it has a close and, in a foreign currency, a rate, to its last close, and must say whether closes or rates are missing when it has none.

## Context

D-61 and D-62 put the model in the reporting currency and start it no earlier than the FX rates; carrying a close and a rate over days without one matches how holdings are valued (valuation.RateOnOrBefore). Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-61]** Both lines are in the reporting currency (D-62). The model's backtest also needs exchange rates into that currency, so it starts no earlier than they do; NBP's rates begin on 2002-01-02.

docs/DESIGN.md#D-61
