---
title: Request Bodies Bounded By Path = true
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
subject_key: core.ports
property_key: request_bodies_bounded_by_path
canonical_key: mcp-kibi-model-requirement:core.ports:request_bodies_bounded_by_path:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2413CF2990A49BB8
claim_text: 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.'
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:54.207Z'
id: FACT-PROP-2609113EE1C0B6CC
type: fact
---
