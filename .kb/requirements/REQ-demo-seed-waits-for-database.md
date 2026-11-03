---
title: The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.
status: open
text_ref: docs/dev/demo.md#demo-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-demo-data
  - lane:strict
semantic_text: The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.
logic_claims:
  - CLAIM-CD04E6DB03EC835A
semantic_clauses:
  - The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a641eec16b876657c2e4056ec90f667689ccb2eef2a505c10ec21b79dcf3dcb0
semantic_inventory:
  - claim_key: CLAIM-CD04E6DB03EC835A
    claim_text: The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound
    role: normative
    status: modeled
    span:
      start: 0
      end: 217
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:15.262Z'
id: REQ-demo-seed-waits-for-database
type: req
---
The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.

## Context

The demo-seed Job starts together with PostgreSQL and the core; an attempt spent on a server still starting brings the Job to its backoff limit. Found by CI run 41676052961 (e2e on k3d): the demo-seed Job and the core migrated River's schema at once and one failed on a duplicate key, and the Job spent its attempts while PostgreSQL was still starting. Modeled by Claude Code at the orchestrator's request; the author has not reviewed the wording.

## Source

> The core and the demo-seed Job start together

deploy/demo/openfief/kustomization.yaml
