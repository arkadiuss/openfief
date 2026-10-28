---
title: The core must answer GET /healthz with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API with 503 until it is ready.
status: open
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
semantic_text: The core must answer GET /healthz with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API with 503 until it is ready.
logic_claims:
  - CLAIM-537D15ACFAE875AC
semantic_clauses:
  - The core must answer GET /healthz with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API with 503 until it is ready.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eb5b843bd389d3d7713184ccdfdc33c0319d4bb114910b31d1e24d63db1375d2
semantic_inventory:
  - claim_key: CLAIM-537D15ACFAE875AC
    claim_text: The core must answer GET /healthz with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API with 503 until it is ready
    role: normative
    status: modeled
    span:
      start: 0
      end: 267
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:08.920Z'
id: REQ-core-health-endpoints
type: req
---
The core must answer GET /healthz with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API with 503 until it is ready.

## Context

Kubernetes probes need endpoints that work without credentials; they must never leak data because the demo exposes the core through an Ingress (D-28 leaves exposure to the installation). Readiness waits for migrations so no request hits a half-built schema. Modeled by Claude Code for the core task: the author asked to see OpenFief's core running on the demo; the author has not reviewed this requirement yet.

## Source

> The core and UI are ClusterIP services; the default way in is `kubectl port-forward`.

docs/DESIGN.md D-28
