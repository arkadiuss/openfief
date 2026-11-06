---
title: Filtered By Bound Accounts = true
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
subject_key: views.sync_history
property_key: filtered_by_bound_accounts
canonical_key: mcp-kibi-model-requirement:views.sync_history:filtered_by_bound_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E75FEAB02F01B017
claim_text: The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer.
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:13.993Z'
id: FACT-PROP-2A448CA391DFDC2D
type: fact
---
