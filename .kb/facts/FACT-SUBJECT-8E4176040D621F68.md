---
title: Valuation Revaluation Trigger
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.revaluation_trigger
canonical_key: valuation.revaluation_trigger
claim_key: CLAIM-027C2A8BCB07FB43
claim_text: 'Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month''s first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date.'
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:24.669Z'
id: FACT-SUBJECT-8E4176040D621F68
type: fact
---
