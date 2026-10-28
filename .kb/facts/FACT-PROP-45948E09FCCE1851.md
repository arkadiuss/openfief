---
title: Coalesces Per Account = true
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
subject_key: valuation.revaluation_queue
property_key: coalesces_per_account
canonical_key: mcp-kibi-model-requirement:valuation.revaluation_queue:coalesces_per_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D9E6D69119281112
claim_text: Revaluations of one account enqueued together, or before one of them runs, must coalesce into a single revaluation from the earliest of their dates.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:41.464Z'
id: FACT-PROP-45948E09FCCE1851
type: fact
---
