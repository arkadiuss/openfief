---
title: New Trace At Edge None Outside = true
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
subject_key: telemetry.traces
property_key: new_trace_at_edge_none_outside
canonical_key: mcp-kibi-model-requirement:telemetry.traces:new_trace_at_edge_none_outside:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AF46315114925B48
claim_text: A request on the core's browser port must start a new trace whatever trace context it carries, and no trace context may be sent outside the cluster.
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:04.316Z'
id: FACT-PROP-E2A6A6ED59BD2E8F
type: fact
---
