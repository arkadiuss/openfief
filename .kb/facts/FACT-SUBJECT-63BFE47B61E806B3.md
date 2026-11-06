---
title: Ingest Isin Lookup
status: active
text_ref: design-decisions:D-166#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-166-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.isin_lookup
canonical_key: ingest.isin_lookup
claim_key: CLAIM-2E620D9AB6C98199
claim_text: 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.'
origin:
  kind: agent
  recorded_at: '2026-11-02T00:55:43.933Z'
id: FACT-SUBJECT-63BFE47B61E806B3
type: fact
---
