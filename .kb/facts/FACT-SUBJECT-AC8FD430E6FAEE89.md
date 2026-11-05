---
title: Ingest Sync Status
status: active
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.sync_status
canonical_key: ingest.sync_status
claim_key: CLAIM-32DF80F61CB1CFF4
claim_text: A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:09.260Z'
id: FACT-SUBJECT-AC8FD430E6FAEE89
type: fact
---
