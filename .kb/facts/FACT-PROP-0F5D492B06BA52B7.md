---
title: Serves Core Until Stopped = true
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.serve_connector
property_key: serves_core_until_stopped
canonical_key: mcp-kibi-model-requirement:sdk.serve_connector:serves_core_until_stopped:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-396DD329FED54C85
claim_text: The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:13.213Z'
id: FACT-PROP-0F5D492B06BA52B7
type: fact
---
