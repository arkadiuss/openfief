---
title: Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction.
status: open
text_ref: design-decisions:D-36#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-3
  - lane:strict
semantic_text: Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction.
logic_claims:
  - CLAIM-55D889CC632D8C71
semantic_clauses:
  - Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 838471cd9278bbfe0aa6b5e083b23ef3f42925f1d061cc3d26a4704ac4392db5
semantic_inventory:
  - claim_key: CLAIM-55D889CC632D8C71
    claim_text: Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 129
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#3
  recorded_at: '2026-10-28T06:29:22.793Z'
id: REQ-decision-d36-3-revaluation-same-transaction
type: req
---
Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#3 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River).

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#3
