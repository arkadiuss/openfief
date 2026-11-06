---
title: Ingest Sync History
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
subject_key: ingest.sync_history
canonical_key: ingest.sync_history
claim_key: CLAIM-0B1092DF451DC233
claim_text: Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:04.019Z'
id: FACT-SUBJECT-8B73C64EB8EB10C2
type: fact
---
