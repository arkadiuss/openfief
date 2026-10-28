---
title: An account's valuation status must give its valued-through date, pending while a change waits, a job waits or a failed run awaits a retry, and running while a job runs.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: An account's valuation status must give its valued-through date, pending while a change waits, a job waits or a failed run awaits a retry, and running while a job runs.
logic_claims:
  - CLAIM-78A42E41215CE251
semantic_clauses:
  - An account's valuation status must give its valued-through date, pending while a change waits, a job waits or a failed run awaits a retry, and running while a job runs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a4432b13ccf2b7500dd9ae995f803a7c16f7332b19d89d4538d7eef6032c6b67
semantic_inventory:
  - claim_key: CLAIM-78A42E41215CE251
    claim_text: An account's valuation status must give its valued-through date, pending while a change waits, a job waits or a failed run awaits a retry, and running while a job runs
    role: normative
    status: modeled
    span:
      start: 0
      end: 167
origin:
  kind: agent
  recorded_at: '2026-10-28T07:58:19.158Z'
id: REQ-valuation-status
type: req
---
An account's valuation status must give its valued-through date, pending while a change waits, a job waits or a failed run awaits a retry, and running while a job runs.

## Context

The read API reports per account the valued-through date and whether a revaluation is pending or running (D-36); this is the query it reads. The read API itself is another task. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it.

docs/DESIGN.md D-36
