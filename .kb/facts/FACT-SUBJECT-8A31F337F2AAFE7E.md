---
title: Ingest Connectors
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.connectors
canonical_key: ingest.connectors
claim_key: CLAIM-9AA347F7A866EFE8
claim_text: 'A sync must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while uploads and bond lots are booked as the owner sent them.'
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:35.665Z'
id: FACT-SUBJECT-8A31F337F2AAFE7E
type: fact
---
