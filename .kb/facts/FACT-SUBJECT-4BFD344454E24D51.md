---
title: Ingest Scheduler
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.scheduler
canonical_key: ingest.scheduler
claim_key: CLAIM-ED7D7D8E5EF86F7D
claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:27.468Z'
id: FACT-SUBJECT-4BFD344454E24D51
type: fact
---
