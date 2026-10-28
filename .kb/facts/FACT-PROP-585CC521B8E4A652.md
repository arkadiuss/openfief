---
title: Own Schema = true
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_store
property_key: own_schema
canonical_key: mcp-kibi-model-requirement:ingest.sync_store:own_schema:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E54E7CB819CE27A7
claim_text: The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:46.345Z'
id: FACT-PROP-585CC521B8E4A652
type: fact
---
