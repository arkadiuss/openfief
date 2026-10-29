---
title: Ingest Ibkrflex Mapper
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.ibkrflex_mapper
canonical_key: ingest.ibkrflex_mapper
claim_key: CLAIM-85C522007389533A
claim_text: The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact.
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:32.954Z'
id: FACT-SUBJECT-5D8CEDAE5085A3D0
type: fact
---
