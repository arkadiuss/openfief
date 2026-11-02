---
title: Core Ports
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: core.ports
canonical_key: core.ports
claim_key: CLAIM-2413CF2990A49BB8
claim_text: 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.'
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:52.130Z'
id: FACT-SUBJECT-377E7BAFECB96D53
type: fact
---
