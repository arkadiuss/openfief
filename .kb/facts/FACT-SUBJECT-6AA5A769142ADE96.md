---
title: Ingest Delivery
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: ingest.delivery
canonical_key: ingest.delivery
claim_key: CLAIM-DCDFF191AB58A6CD
claim_text: A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:03.387Z'
id: FACT-SUBJECT-6AA5A769142ADE96
type: fact
---
