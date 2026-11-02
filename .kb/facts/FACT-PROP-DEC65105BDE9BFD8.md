---
title: Request Bodies Bounded Before Auth = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.read_api
property_key: request_bodies_bounded_before_auth
canonical_key: mcp-kibi-model-requirement:views.read_api:request_bodies_bounded_before_auth:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5F9952885EB49774
claim_text: The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message.
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:39.878Z'
id: FACT-PROP-DEC65105BDE9BFD8
type: fact
---
