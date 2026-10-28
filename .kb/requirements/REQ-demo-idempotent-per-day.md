---
title: Loading the same seed and day's dataset into a demo database again must change nothing, and loading another day's dataset without --reset-demo must fail with an error that asks for --reset-demo.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Loading the same seed and day's dataset into a demo database again must change nothing, and loading another day's dataset without --reset-demo must fail with an error that asks for --reset-demo.
logic_claims:
  - CLAIM-1D2F5FFD344B4805
semantic_clauses:
  - Loading the same seed and day's dataset into a demo database again must change nothing, and loading another day's dataset without --reset-demo must fail with an error that asks for --reset-demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f0d22c79744ec61122115843d4affcf6d2e88bbb476080ea983206079dd6a8c9
semantic_inventory:
  - claim_key: CLAIM-1D2F5FFD344B4805
    claim_text: Loading the same seed and day's dataset into a demo database again must change nothing, and loading another day's dataset without --reset-demo must fail with an error that asks for --reset-demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 193
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:47.998Z'
id: REQ-demo-idempotent-per-day
type: req
---
Loading the same seed and day's dataset into a demo database again must change nothing, and loading another day's dataset without --reset-demo must fail with an error that asks for --reset-demo.

## Context

A Kubernetes Job may run the loader more than once; every import is idempotent by key (D-18) and a seeding is refused once an account is seeded (D-23), so a re-run the same day is a no-op and another day's go-live needs an explicit reset. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> Re-imports are idempotent, keyed by source plus the institution's own ID

docs/DESIGN.md D-18
