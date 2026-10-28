---
title: Ingest Ingest Path
status: active
text_ref: docs/DESIGN.md#D-7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-7
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.ingest_path
canonical_key: ingest.ingest_path
claim_key: CLAIM-4007313BD110B132
claim_text: A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:37:24.167Z'
id: FACT-SUBJECT-EF78119FF140D389
type: fact
---
