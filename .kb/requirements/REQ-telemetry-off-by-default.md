---
title: 'Trace export must stay off unless an OTLP endpoint is configured: without one no span is recorded and no connection is opened.'
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: 'Trace export must stay off unless an OTLP endpoint is configured: without one no span is recorded and no connection is opened.'
logic_claims:
  - CLAIM-D7B46FD354B74406
semantic_clauses:
  - 'Trace export must stay off unless an OTLP endpoint is configured: without one no span is recorded and no connection is opened.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 305fa229cc30331be5834aa1543dea12ed3e5e2df430a401219a4942092004e1
semantic_inventory:
  - claim_key: CLAIM-D7B46FD354B74406
    claim_text: 'Trace export must stay off unless an OTLP endpoint is configured: without one no span is recorded and no connection is opened'
    role: exception
    status: modeled
    span:
      start: 0
      end: 125
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:12.884Z'
id: REQ-telemetry-off-by-default
type: req
---
Trace export must stay off unless an OTLP endpoint is configured: without one no span is recorded and no connection is opened.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. "Configurable" is read as off by default: an installation that runs no collector pays nothing and sends nothing. telemetry.Start installs a no-op tracer provider then; OTEL_SDK_DISABLED=true and OTEL_TRACES_EXPORTER=none turn an endpoint off too.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
