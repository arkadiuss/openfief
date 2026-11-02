---
title: Ingest Ticker Service
status: active
text_ref: docs/DESIGN.md#D-153
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-153
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.ticker_service
canonical_key: ingest.ticker_service
claim_key: CLAIM-38F3CE700270FE8D
claim_text: The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:46.594Z'
id: FACT-SUBJECT-51265BDA30800E3F
type: fact
---
