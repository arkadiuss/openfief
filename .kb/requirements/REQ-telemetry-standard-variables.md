---
title: 'Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes.'
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: 'Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes.'
logic_claims:
  - CLAIM-C234BAF628D992EB
semantic_clauses:
  - 'Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 484ad459adf7ed42726fb057a9c71c424ae4742f23977fd660df19f11a58419b
semantic_inventory:
  - claim_key: CLAIM-C234BAF628D992EB
    claim_text: 'Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes'
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:30.009Z'
id: REQ-telemetry-standard-variables
type: req
---
Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. The standard variables make any OTLP collector work (the OpenTelemetry Collector, Tempo, Jaeger) and are what operators already know. telemetry.ParseConfig reads them; the traces-specific variant of each wins.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
