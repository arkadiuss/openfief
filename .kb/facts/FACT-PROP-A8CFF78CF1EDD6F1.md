---
title: Standard Environment Variables = true
status: active
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: telemetry.export
property_key: standard_environment_variables
canonical_key: mcp-kibi-model-requirement:telemetry.export:standard_environment_variables:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C234BAF628D992EB
claim_text: 'Trace export must be configured with the standard OpenTelemetry environment variables: the OTLP endpoint, the protocol (grpc or http/protobuf), headers, the sampler and its argument, the service name and resource attributes.'
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:27.262Z'
id: FACT-PROP-A8CFF78CF1EDD6F1
type: fact
---
