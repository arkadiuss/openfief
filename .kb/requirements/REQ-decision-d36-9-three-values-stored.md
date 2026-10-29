---
title: 'Each stored valuation must hold all three values of D-77: gross, net and liquidation.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-9
  - provenance:design-decisions-d-77
  - lane:strict
text_ref: design-decisions:D-36#9
origin:
  kind: import
  ref: design-decisions:D-36#9
  recorded_at: '2026-10-29T08:06:09.530Z'
semantic_text: 'Each stored valuation must hold all three values of D-77: gross, net and liquidation.'
logic_claims:
  - CLAIM-4435567D45B07F9D
semantic_clauses:
  - 'Each stored valuation must hold all three values of D-77: gross, net and liquidation.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d317e7b2c88f76262a7a45349a2dd3f829bffdd33aa77f4f7d7962906f8e3385
semantic_inventory:
  - claim_key: CLAIM-4435567D45B07F9D
    claim_text: 'Each stored valuation must hold all three values of D-77: gross, net and liquidation'
    role: normative
    status: modeled
    span:
      start: 0
      end: 84
id: REQ-decision-d36-9-three-values-stored
type: req
---
Each stored valuation must hold all three values of D-77: gross, net and liquidation.

## Why this changed

The author's decision D-77 gives every holding, account and total three values instead of D-25's single configurable bond basis, and D-36's text changed with it: each stored valuation now holds all three values, and the sentence that changing the basis revalues everything is gone, since there is no basis left to change. REQ-decision-d36-9-basis-change-revalues-all recorded that sentence, so this requirement, D-36's new clause in its place, replaces it.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#9 as a firm rule, so changing it needs the author. Reason not stated in the decision beyond D-77. REQ-valuation-three-values, REQ-valuation-gross-value, REQ-valuation-net-value and REQ-valuation-liquidation-value record D-77; REQ-decision-d36-8-concurrency-and-schedule records D-36's new settings. The implementation still stores one value per holding and day, keyed by a configured basis (REQ-valuation-revaluation-rows, REQ-valuation-basis-change). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date. The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it. Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-36#9
