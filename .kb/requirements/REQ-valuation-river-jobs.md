---
title: Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.
logic_claims:
  - CLAIM-8CBD69ADFDF487D2
semantic_clauses:
  - Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a2f7c9294aa10ec3b4a6240cc79abce1f2242ac6d33036ffde1e895739fc86d6
semantic_inventory:
  - claim_key: CLAIM-8CBD69ADFDF487D2
    claim_text: Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job
    role: normative
    status: modeled
    span:
      start: 0
      end: 279
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:47.493Z'
id: REQ-valuation-river-jobs
type: req
---
Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.

## Context

River dedupes unique jobs by their args and must count running jobs among unique states, so the date cannot live in the args; it lives in valuation.dirty and every enqueue lowers it with LEAST(). Completing the job in the transaction that clears the date, under a per-account advisory lock that enqueueing also takes, closes the race with a change during a run. Refines REQ-decision-d36-4-river-job-queue and REQ-decision-d36-5-revaluation-coalescing. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date.

docs/DESIGN.md D-36
