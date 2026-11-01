---
title: Ingest Sync Request
status: active
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.sync_request
canonical_key: ingest.sync_request
claim_key: CLAIM-4512B4382550D09D
claim_text: Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:41.020Z'
id: FACT-SUBJECT-5769DD481E8026B5
type: fact
---
