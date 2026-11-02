---
title: Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint.
logic_claims:
  - CLAIM-F2499E1C6E6CB495
semantic_clauses:
  - Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7cfcdba020ff020a4d4a7ed6f66d6dd59134aeea766f342dab533ee88d1d4606
semantic_inventory:
  - claim_key: CLAIM-F2499E1C6E6CB495
    claim_text: Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint
    role: exception
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:35.106Z'
id: REQ-telemetry-plugins-sdk
type: req
---
Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. First-party plugins should get tracing for free (D-11: the SDK is what plugins are built on). The operator does not pass the variables to plugins and their NetworkPolicy and the egress proxy keep them from a collector, so today a plugin only carries the trace through (implementation question 71).

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
