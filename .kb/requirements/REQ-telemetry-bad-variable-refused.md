---
title: A tracing variable whose value cannot be used must stop the process at start with an error that names the variable and never quotes an endpoint or a header.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: A tracing variable whose value cannot be used must stop the process at start with an error that names the variable and never quotes an endpoint or a header.
logic_claims:
  - CLAIM-73D3ED99EBC6B2E3
semantic_clauses:
  - A tracing variable whose value cannot be used must stop the process at start with an error that names the variable and never quotes an endpoint or a header.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dae0a7e1b1deb7c5676103f25bfb8134b06b45adbb9ede5cc59ff200cf9c2cc6
semantic_inventory:
  - claim_key: CLAIM-73D3ED99EBC6B2E3
    claim_text: A tracing variable whose value cannot be used must stop the process at start with an error that names the variable and never quotes an endpoint or a header
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:38.666Z'
id: REQ-telemetry-bad-variable-refused
type: req
---
A tracing variable whose value cannot be used must stop the process at start with an error that names the variable and never quotes an endpoint or a header.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. A typo must not silently turn tracing off or send spans somewhere else, and an endpoint or header may hold a credential (GR-12), so the error names the variable only. Values are checked whether export is on or not.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
