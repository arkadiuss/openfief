---
title: Propagated Across Processes And Jobs = true
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
property_key: propagated_across_processes_and_jobs
canonical_key: mcp-kibi-model-requirement:telemetry.traces:propagated_across_processes_and_jobs:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CC1B29495EE9BE4A
claim_text: A trace must continue across the core's calls to plugins and plugins' calls to the core in W3C trace context, and a revaluation job's work must link to the span that queued it.
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:55.377Z'
id: FACT-PROP-56EA30A301E8CC22
type: fact
---
