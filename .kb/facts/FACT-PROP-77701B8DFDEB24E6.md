---
title: Counts Per Viewable Account = true
status: active
text_ref: docs/DESIGN.md#D-199
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-199
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.activity
property_key: counts_per_viewable_account
canonical_key: mcp-kibi-model-requirement:views.activity:counts_per_viewable_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F8784D5E9806C93C
claim_text: An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account.
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:19.296Z'
id: FACT-PROP-77701B8DFDEB24E6
type: fact
---
