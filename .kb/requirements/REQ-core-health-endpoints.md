---
title: The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready.
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready.
logic_claims:
  - CLAIM-E66A9980ABD5F2D2
semantic_clauses:
  - The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 20c254e8e71f217f4224fb87457798f0b002adbd3a253ea8187ab818c8483fd9
semantic_inventory:
  - claim_key: CLAIM-E66A9980ABD5F2D2
    claim_text: The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready
    role: normative
    status: modeled
    span:
      start: 0
      end: 332
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:08.920Z'
id: REQ-core-health-endpoints
type: req
---
The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready.

## Context

Kubernetes probes need liveness at once and readiness only once the core can serve; neither may leak data. D-115 puts the health endpoints on the browser port only; the plugin port has none, and both ports answer 503 until the core is wired, so nothing half-started serves a plugin or a browser.

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

docs/DESIGN.md#D-115
