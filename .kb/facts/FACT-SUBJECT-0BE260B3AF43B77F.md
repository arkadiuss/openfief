---
title: Identity Script Token Store
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.script_token_store
canonical_key: identity.script_token_store
claim_key: CLAIM-95DD9512447C8C06
claim_text: The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:33.798Z'
id: FACT-SUBJECT-0BE260B3AF43B77F
type: fact
---
