---
title: Health Endpoints = true
status: active
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: core.service
property_key: health_endpoints
canonical_key: mcp-kibi-model-requirement:core.service:health_endpoints:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E66A9980ABD5F2D2
claim_text: The core must answer GET /healthz on its browser port with 200 as soon as it listens and GET /readyz with 200 only once it is migrated, not shutting down and its database answers, both without a token, without data and uncacheable, and must answer the read API on both ports and the plugin port's services with 503 until it is ready.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:06.782Z'
id: FACT-PROP-D54777BA6F452EA6
type: fact
---
