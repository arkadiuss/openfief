---
title: Ingest Sync Retry
status: active
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.sync_retry
canonical_key: ingest.sync_retry
claim_key: CLAIM-F697BAF8E0B32439
claim_text: The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.
origin:
  kind: agent
  recorded_at: '2026-11-06T04:22:59.638Z'
id: FACT-SUBJECT-E4D322C8278B6635
type: fact
---
