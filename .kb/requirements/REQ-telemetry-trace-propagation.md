---
title: A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it.
logic_claims:
  - CLAIM-CC1B29495EE9BE4A
semantic_clauses:
  - A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a421466a384dc253fd7a7d56f94b351a65c2a7ba4765f6e63c880cc82dd72beb
semantic_inventory:
  - claim_key: CLAIM-CC1B29495EE9BE4A
    claim_text: A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it
    role: normative
    status: modeled
    span:
      start: 0
      end: 175
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:57.957Z'
id: REQ-telemetry-trace-propagation
type: req
---
A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. One sync spans the core, a connector plugin, the ledger and the revaluation queue (D-36, D-49); a trace is only useful if it follows that path. The core's connector client, the front door and the SDK's clients inject traceparent; the plugin port and the SDK's server continue it; River jobs carry it in their metadata.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
