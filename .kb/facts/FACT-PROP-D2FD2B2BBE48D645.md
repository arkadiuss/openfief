---
title: Validates Targets And Weights = true
status: active
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: portfolio.model_portfolio
property_key: validates_targets_and_weights
canonical_key: mcp-kibi-model-requirement:portfolio.model_portfolio:validates_targets_and_weights:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2D0B55914D301D47
claim_text: A model portfolio target must be an instrument named by ISIN and venue symbol or an index with its currency, named once, weighted above 0% and at most 100% in steps of 0.01%; a model may have at most 50 targets whose weights sum to exactly 100%, and may be on only with targets.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:33.830Z'
id: FACT-PROP-D2FD2B2BBE48D645
type: fact
---
