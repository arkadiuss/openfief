---
title: Ingest Ticker Confirmation
status: active
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.ticker_confirmation
canonical_key: ingest.ticker_confirmation
claim_key: CLAIM-5491D9D37B051211
claim_text: 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting.'
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:11.659Z'
id: FACT-SUBJECT-68A4C70B5198225B
type: fact
---
