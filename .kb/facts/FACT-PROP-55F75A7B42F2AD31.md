---
title: Answer Counted Before Decoding = true
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
property_key: answer_counted_before_decoding
canonical_key: mcp-kibi-model-requirement:ingest.connector_client:answer_counted_before_decoding:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4425DC9262ED8418
claim_text: The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it.
origin:
  kind: agent
  recorded_at: '2026-11-03T05:55:08.800Z'
id: FACT-PROP-55F75A7B42F2AD31
type: fact
---
