---
title: CI must run the full task check on every push to main and to the implementation branch, and only main may publish the docs, push images or deploy the demo.
status: open
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
semantic_text: CI must run the full task check on every push to main and to the implementation branch, and only main may publish the docs, push images or deploy the demo.
logic_claims:
  - CLAIM-C0CFC147A698A3EE
semantic_clauses:
  - CI must run the full task check on every push to main and to the implementation branch, and only main may publish the docs, push images or deploy the demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2d30ce7eec3a964df1e8620e91aa9ed12b030deca6a93258a4e2072eb084ddef
semantic_inventory:
  - claim_key: CLAIM-C0CFC147A698A3EE
    claim_text: CI must run the full task check on every push to main and to the implementation branch, and only main may publish the docs, push images or deploy the demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 154
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:51.990Z'
id: REQ-tooling-ci-implementation-branch
type: req
---
CI must run the full task check on every push to main and to the implementation branch, and only main may publish the docs, push images or deploy the demo.

## Context

With the hook as a fast tier, main advances only to a commit on which CI's full check is green, so the branch the orchestrator pushes needs the same run a push to main gets, without publishing anything. The author approved the fast-gate plan (the pre-commit hook as a fast tier, the full gate in CI, retried downloads); implemented by Claude Code on the build/fast-gate branch, and the author has not reviewed this wording yet. 

## Source

> `main` advances only after CI is green on the commit it moves to: a pull request merges when its run is green, and the implementation branch is fast-forwarded into `main` once its own run is green on that exact commit.

docs/dev/sdlc.md#gate-tiers
