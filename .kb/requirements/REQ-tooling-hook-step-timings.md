---
title: The pre-commit hook must print how long each step took and the total, also when a step fails.
status: open
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
semantic_text: The pre-commit hook must print how long each step took and the total, also when a step fails.
logic_claims:
  - CLAIM-22A751E3500729C4
semantic_clauses:
  - The pre-commit hook must print how long each step took and the total, also when a step fails.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 57c505c23dda04acf4a0698813350efd3d7c63353212c05aa1ec32353280f5c6
semantic_inventory:
  - claim_key: CLAIM-22A751E3500729C4
    claim_text: The pre-commit hook must print how long each step took and the total, also when a step fails
    role: normative
    status: modeled
    span:
      start: 0
      end: 92
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:47.182Z'
id: REQ-tooling-hook-step-timings
type: req
---
The pre-commit hook must print how long each step took and the total, also when a step fails.

## Context

The fast-gate plan sets a target of a median of 5 minutes and a p95 of 8 for the hook, and the author can only keep it if each commit shows where its time went. The author approved the fast-gate plan (the pre-commit hook as a fast tier, the full gate in CI, retried downloads); implemented by Claude Code on the build/fast-gate branch, and the author has not reviewed this wording yet. 

## Source

> the script runs them in order, prints each one's time (`check-commit: <task>: 41s`) and the total, and stops at the first failure.

docs/dev/sdlc.md#gate-tiers
