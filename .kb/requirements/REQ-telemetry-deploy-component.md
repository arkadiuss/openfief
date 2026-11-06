---
title: The install bases must keep tracing off, and an optional kustomize component must set the tracing variables for the core and the operator and let the core reach the collector.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: The install bases must keep tracing off, and an optional kustomize component must set the tracing variables for the core and the operator and let the core reach the collector.
logic_claims:
  - CLAIM-2DDDC65270A656B6
semantic_clauses:
  - The install bases must keep tracing off, and an optional kustomize component must set the tracing variables for the core and the operator and let the core reach the collector.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5fd6c1494eba366b55e58965ef804c8988ee85f930b8bba28b5ce5091fa827d3
semantic_inventory:
  - claim_key: CLAIM-2DDDC65270A656B6
    claim_text: The install bases must keep tracing off, and an optional kustomize component must set the tracing variables for the core and the operator and let the core reach the collector
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:46.225Z'
id: REQ-telemetry-deploy-component
type: req
---
The install bases must keep tracing off, and an optional kustomize component must set the tracing variables for the core and the operator and let the core reach the collector.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. The core's NetworkPolicy allows only DNS, the database, plugins and the API server (D-92), so export needs a rule; like deploy/gateway (D-28) it lives in a component outside the bases. The collector's credential comes from a Secret, never a literal.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
