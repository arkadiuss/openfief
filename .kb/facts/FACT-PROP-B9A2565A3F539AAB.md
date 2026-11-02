---
title: Opening Upload Keeps Started Accounts Out = true
status: active
text_ref: docs/DESIGN.md#D-140
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-140
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.go_live
property_key: opening_upload_keeps_started_accounts_out
canonical_key: mcp-kibi-model-requirement:ingest.go_live:opening_upload_keeps_started_accounts_out:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0288B79F8E43FFDB
claim_text: 'A sync, or a statement uploaded in transactions mode, must not book again what an account''s seeding holds: an opening batch must leave out the seeding of an account already seeded, and a transactions batch the transactions and ticker entries of a seeded account dated on or before its go-live date, each as a skipped ingest problem; an opening upload must leave out the seeding of the other bound accounts that have already started and keep the account it was uploaded for as sent, and bond lots are booked as the owner sent them.'
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:53.337Z'
id: FACT-PROP-B9A2565A3F539AAB
type: fact
---
