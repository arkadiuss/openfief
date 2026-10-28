---
title: Revaluations must run through a PostgreSQL-backed job queue (River).
status: open
text_ref: design-decisions:D-36#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-4
  - lane:strict
semantic_text: Revaluations must run through a PostgreSQL-backed job queue (River).
logic_claims:
  - CLAIM-0B4C3DA403CCB616
semantic_clauses:
  - Revaluations must run through a PostgreSQL-backed job queue (River).
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b8e5ef0433c746260fb018057811e951fed3b6d2ccb5aa263052aba628c4e98f
semantic_inventory:
  - claim_key: CLAIM-0B4C3DA403CCB616
    claim_text: Revaluations must run through a PostgreSQL-backed job queue (River)
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#4
  recorded_at: '2026-10-28T06:29:33.459Z'
id: REQ-decision-d36-4-river-job-queue
type: req
---
Revaluations must run through a PostgreSQL-backed job queue (River).

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#4 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River).

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#4
