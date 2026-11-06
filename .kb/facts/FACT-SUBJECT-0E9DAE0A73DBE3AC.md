---
title: Attention Failing Sync Check
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
subject_key: attention.failing_sync_check
canonical_key: attention.failing_sync_check
claim_key: CLAIM-ECCBD4961D87307A
claim_text: The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:52.737Z'
id: FACT-SUBJECT-0E9DAE0A73DBE3AC
type: fact
---
