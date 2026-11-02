---
title: Ingest Connector Client
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
subject_key: ingest.connector_client
canonical_key: ingest.connector_client
claim_key: CLAIM-9E2E735844C765DF
claim_text: The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.
origin:
  kind: agent
  recorded_at: '2026-11-02T19:46:00.510Z'
id: FACT-SUBJECT-48574C3BDC2B27C0
type: fact
---
