---
title: A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster.
logic_claims:
  - CLAIM-AF46315114925B48
semantic_clauses:
  - A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9fc156523001dbeb6f5fe145416adae49692902b4e607476dd9088cdda61a546
semantic_inventory:
  - claim_key: CLAIM-AF46315114925B48
    claim_text: A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:06.977Z'
id: REQ-telemetry-trace-boundary
type: req
---
A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. A caller that could choose the trace could force sampling or join its own trace to the owner's; the front door (D-81) also drops a browser's trace headers before proxying to the view plugin. Calls to institutions through the egress proxy carry no trace context and are not traced (implementation question 72).

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
