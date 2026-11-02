---
title: Sdk Go Telemetry
status: active
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: sdk_go.telemetry
canonical_key: sdk_go.telemetry
claim_key: CLAIM-F2499E1C6E6CB495
claim_text: Plugins built on the Go SDK must trace their connector calls and their calls to the core without code of their own, and export nothing unless their environment names an OTLP endpoint.
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:30.126Z'
id: FACT-SUBJECT-017BD07A11139A56
type: fact
---
