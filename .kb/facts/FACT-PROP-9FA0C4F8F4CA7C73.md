---
title: Records Missing Input = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.holding_valuation
property_key: records_missing_input
canonical_key: mcp-kibi-model-requirement:valuation.holding_valuation:records_missing_input:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FBF72D109CB01BB4
claim_text: 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.'
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:27.430Z'
id: FACT-PROP-9FA0C4F8F4CA7C73
type: fact
---
