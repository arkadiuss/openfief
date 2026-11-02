---
title: On shutdown a process must send the spans it still buffers, waiting a bounded time.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: On shutdown a process must send the spans it still buffers, waiting a bounded time.
logic_claims:
  - CLAIM-2B8E99ADE8537BF5
semantic_clauses:
  - On shutdown a process must send the spans it still buffers, waiting a bounded time.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 026cbcf0884ebf0c4484bf413a06b90394e6e078aca48b34b041a92a00adb428
semantic_inventory:
  - claim_key: CLAIM-2B8E99ADE8537BF5
    claim_text: On shutdown a process must send the spans it still buffers, waiting a bounded time
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:47.342Z'
id: REQ-telemetry-flush-on-shutdown
type: req
---
On shutdown a process must send the spans it still buffers, waiting a bounded time.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. Spans are exported in batches; without a flush the last operations before a restart, often the interesting ones, would be lost. The core and the operator wait up to five seconds; a plugin up to its shutdown timeout.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
