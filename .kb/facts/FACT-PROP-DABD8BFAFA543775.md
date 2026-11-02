---
title: Answer Bounded = true
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
subject_key: ingest.connector_client
property_key: answer_bounded
canonical_key: mcp-kibi-model-requirement:ingest.connector_client:answer_bounded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9E2E735844C765DF
claim_text: The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.
origin:
  kind: agent
  recorded_at: '2026-11-02T19:46:02.768Z'
id: FACT-PROP-DABD8BFAFA543775
type: fact
---
