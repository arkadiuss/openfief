---
title: Ingest Go Live
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
subject_key: ingest.go_live
canonical_key: ingest.go_live
claim_key: CLAIM-CCA301C6324310BF
claim_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.'
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:27.296Z'
id: FACT-SUBJECT-C256554D473F5098
type: fact
---
