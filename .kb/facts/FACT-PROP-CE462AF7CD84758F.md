---
title: Sync Keeps From Go Live = true
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connectors
property_key: sync_keeps_from_go_live
canonical_key: mcp-kibi-model-requirement:ingest.connectors:sync_keeps_from_go_live:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9AA347F7A866EFE8
claim_text: 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.'
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:38.424Z'
id: FACT-PROP-CE462AF7CD84758F
type: fact
---
