---
title: Answer From Ledger Queries = true
status: active
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.queries
property_key: answer_from_ledger_queries
canonical_key: mcp-kibi-model-requirement:views.queries:answer_from_ledger_queries:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6B5ED0620547FE66
claim_text: The read API's per-account queries must answer from the ledger's queries over the current transaction log, must refuse an account that is not stored with an unknown account error, and must report a store failure as a failure rather than as an unknown account.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:47.058Z'
id: FACT-PROP-C03FFCE5535D278E
type: fact
---
