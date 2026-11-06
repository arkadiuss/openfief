---
title: Lists Items By View With Group And Scope = true
status: active
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.attention_inbox
property_key: lists_items_by_view_with_group_and_scope
canonical_key: mcp-kibi-model-requirement:views.attention_inbox:lists_items_by_view_with_group_and_scope:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-738581D03517F495
claim_text: The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view.
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:31.636Z'
id: FACT-PROP-A88C7C387F9376E9
type: fact
---
