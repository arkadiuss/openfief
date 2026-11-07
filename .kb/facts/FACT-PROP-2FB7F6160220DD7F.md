---
title: Lists Latest Unsupported Rows = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.balance_check
property_key: lists_latest_unsupported_rows
canonical_key: mcp-kibi-model-requirement:views.balance_check:lists_latest_unsupported_rows:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A6F4BC94BBFA47A0
claim_text: The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked.
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:24.517Z'
id: FACT-PROP-2FB7F6160220DD7F
type: fact
---
