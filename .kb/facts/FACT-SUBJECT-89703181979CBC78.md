---
title: Ingest Payout Sync
status: active
text_ref: design-decisions:D-162#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-162-3
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.payout_sync
canonical_key: ingest.payout_sync
claim_key: CLAIM-B2544C5D2AF4F210
claim_text: After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.
origin:
  kind: agent
  recorded_at: '2026-11-02T00:55:24.475Z'
id: FACT-SUBJECT-89703181979CBC78
type: fact
---
