---
title: Worker concurrency and the end-of-day schedule must be configurable.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-8
  - provenance:design-decisions-d-77
  - lane:strict
text_ref: design-decisions:D-36#8
origin:
  kind: import
  ref: design-decisions:D-36#8
  recorded_at: '2026-10-29T08:05:23.601Z'
semantic_text: Worker concurrency and the end-of-day schedule must be configurable.
logic_claims:
  - CLAIM-D823A1E694E76D74
semantic_clauses:
  - Worker concurrency and the end-of-day schedule must be configurable.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 556cba7c6e835453d3789b5b697a1fb77ac776df2de75eea2b73ba82ca0c0ed2
semantic_inventory:
  - claim_key: CLAIM-D823A1E694E76D74
    claim_text: Worker concurrency and the end-of-day schedule must be configurable
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
id: REQ-decision-d36-8-concurrency-and-schedule
type: req
---
Worker concurrency and the end-of-day schedule must be configurable.

## Why this changed

The author's decision D-77 replaces D-25's single configurable bond basis with three values, gross, net and liquidation, that every holding, account and total has, and D-36's text changed with it: the bond basis is no longer among D-36's settings, and each stored valuation holds all three values instead. REQ-decision-d36-8-revaluation-settings listed the bond basis as a setting, so this requirement replaces it with D-36's new list.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#8 as a firm rule, so changing it needs the author. Reason not stated in the decision. REQ-decision-d36-9-three-values-stored records D-36's new clause on the three values. The implementation still reads a bond basis setting (REQ-valuation-settings, REQ-core-config). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date. The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it. Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-36#8
