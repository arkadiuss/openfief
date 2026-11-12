---
title: CI must cache the pinned tools by the versions the Taskfile pins, so that a run whose pinned versions did not change installs none of them again.
status: open
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
semantic_text: CI must cache the pinned tools by the versions the Taskfile pins, so that a run whose pinned versions did not change installs none of them again.
logic_claims:
  - CLAIM-4D6E370FE31AA69F
semantic_clauses:
  - CI must cache the pinned tools by the versions the Taskfile pins, so that a run whose pinned versions did not change installs none of them again.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e2d4a4e033a42bf3f527f87ad9406f503827eecd5c467a0c7417fa2828ecf69
semantic_inventory:
  - claim_key: CLAIM-4D6E370FE31AA69F
    claim_text: CI must cache the pinned tools by the versions the Taskfile pins, so that a run whose pinned versions did not change installs none of them again
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
origin:
  kind: agent
  recorded_at: '2026-11-12T15:12:02.944Z'
id: REQ-tooling-tool-cache
type: req
---
CI must cache the pinned tools by the versions the Taskfile pins, so that a run whose pinned versions did not change installs none of them again.

## Context

Every CI job installed the tools again, a large share of its time and of its exposure to a download outage; the tool targets already skip an install whose binary is at the pinned version, so restoring the directory by the pinned versions is all that is missing. The author approved the fast-gate plan (the pre-commit hook as a fast tier, the full gate in CI, retried downloads); implemented by Claude Code on the build/fast-gate branch, and the author has not reviewed this wording yet. 

## Source

> The setup action caches `.bin/` (without `.bin/e2e`, the e2e job's own state) with `actions/cache`, keyed by runner, job and a hash of the Taskfile's `*_VERSION` and checksum variables plus `go.mod`'s `go` and `toolchain` lines.

docs/dev/sdlc.md#gate-tiers
