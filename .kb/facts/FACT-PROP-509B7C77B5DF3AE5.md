---
title: Pages With Opaque Cursor = true
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
subject_key: views.transaction_listing
property_key: pages_with_opaque_cursor
canonical_key: mcp-kibi-model-requirement:views.transaction_listing:pages_with_opaque_cursor:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6F18288550BE7608
claim_text: Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:36.229Z'
id: FACT-PROP-509B7C77B5DF3AE5
type: fact
---
