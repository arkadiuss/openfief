---
title: Demo Seed Database Wait
status: active
text_ref: docs/dev/demo.md#demo-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-demo-data
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: demo.seed_database_wait
canonical_key: demo.seed_database_wait
claim_key: CLAIM-CD04E6DB03EC835A
claim_text: The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:11.375Z'
id: FACT-SUBJECT-8FB08B0A7E84540E
type: fact
---
