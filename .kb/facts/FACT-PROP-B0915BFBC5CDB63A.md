---
title: Reread On Every Call = true
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.go.plugin_token
property_key: reread_on_every_call
canonical_key: mcp-kibi-model-requirement:sdk.go.plugin_token:reread_on_every_call:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4779E08C64B66221
claim_text: The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:12.325Z'
id: FACT-PROP-B0915BFBC5CDB63A
type: fact
---
