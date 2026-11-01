---
title: Later Statements From Go Live = true
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
subject_key: ingest.go_live
property_key: later_statements_from_go_live
canonical_key: mcp-kibi-model-requirement:ingest.go_live:later_statements_from_go_live:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CCA301C6324310BF
claim_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem, while opening uploads and bond lots are booked as the owner sent them.'
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:29.899Z'
id: FACT-PROP-24A69260C96A0EBC
type: fact
---
