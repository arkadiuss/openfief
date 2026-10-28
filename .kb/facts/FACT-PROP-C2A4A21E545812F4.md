---
title: Resolves Range And Revaluation State = true
status: active
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.performance
property_key: resolves_range_and_revaluation_state
canonical_key: mcp-kibi-model-requirement:views.performance:resolves_range_and_revaluation_state:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8CE6227208D074E7
claim_text: The performance query must start at go-live when no start date or an earlier one is given, end by default at the latest valued-through date of any account, count an account with transactions that was never valued as valued through the day before its first transaction, and report whether any revaluation is pending or running.
origin:
  kind: agent
  recorded_at: '2026-10-28T20:55:17.685Z'
id: FACT-PROP-C2A4A21E545812F4
type: fact
---
