---
title: Ingest Sync Request
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-117-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-117#1
fact_kind: subject
subject_key: ingest.sync_request
canonical_key: ingest.sync_request
claim_key: CLAIM-1A300E8ADF318010
claim_text: Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:11:45.961Z'
id: FACT-SUBJECT-8EC6E0F87B5129F6
type: fact
---
The sync request the core sends to a connector instance bound to accounts: for each bound account once in ID order it carries the go-live date, tax wrapper, non-zero holdings and open lots, the stored payout settings and history, the bond series terms and the provisional names. It does not cover the connector's delivery.