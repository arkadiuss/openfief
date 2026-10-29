---
title: Ingest Sync Store
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.sync_store
canonical_key: ingest.sync_store
claim_key: CLAIM-E54E7CB819CE27A7
claim_text: The ingest context must keep each connector instance's watermark, sync state and claim, its import batch history with each batch's state, attempts and counts, and its connector bindings in its own PostgreSQL schema, migrated by its own runner and lock, holding internal IDs, counts and times only.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:44.582Z'
id: FACT-SUBJECT-019F6B72F4372DE8
type: fact
---
