---
title: Change Revalues = true
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
subject_key: valuation.source_ranking
property_key: change_revalues
canonical_key: mcp-kibi-model-requirement:valuation.source_ranking:change_revalues:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9C6C219D57E65B26
claim_text: Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:11.995Z'
id: FACT-PROP-C75D03DACBACADA3
type: fact
---
