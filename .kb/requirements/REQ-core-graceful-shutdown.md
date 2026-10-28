---
title: On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it.
logic_claims:
  - CLAIM-C8298DEF301B39E0
semantic_clauses:
  - On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 986e63e9e19861166ee67105986f13f131041eb89a56302aff3b28fcb02fcd8b
semantic_inventory:
  - claim_key: CLAIM-C8298DEF301B39E0
    claim_text: On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it
    role: normative
    status: modeled
    span:
      start: 0
      end: 175
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:13.725Z'
id: REQ-core-graceful-shutdown
type: req
---
On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it.

## Context

A revaluation cut off mid-transaction is rolled back and retried by River, but draining first avoids needless work and failed requests on every rollout. Reason for the timeout: the pod's termination grace period bounds it. Modeled by Claude Code for the core task: the author asked to see OpenFief's core running on the demo; the author has not reviewed this requirement yet.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River).

docs/DESIGN.md D-36
