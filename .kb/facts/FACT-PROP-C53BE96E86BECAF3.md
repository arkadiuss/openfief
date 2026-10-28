---
title: Queries Use Current Versions = true
status: active
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.derivation
property_key: queries_use_current_versions
canonical_key: mcp-kibi-model-requirement:ledger.derivation:queries_use_current_versions:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-78198D70FFA8D894
claim_text: The ledger must answer an account's current transactions, holdings, lots and cash on a date from its stored log using only current versions, and must refuse a query for an account that is not stored.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:52.359Z'
id: FACT-PROP-C53BE96E86BECAF3
type: fact
---
