---
title: Valuation Holding Valuation
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
subject_key: valuation.holding_valuation
canonical_key: valuation.holding_valuation
claim_key: CLAIM-FBF72D109CB01BB4
claim_text: 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.'
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:22.835Z'
id: FACT-SUBJECT-DB1A2B3E4A183328
type: fact
---
