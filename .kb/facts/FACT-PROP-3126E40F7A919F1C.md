---
title: One Append Only Store With Kind Subject State History = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-130#1
fact_kind: property_value
subject_key: core.attention_items
property_key: one_append_only_store_with_kind_subject_state_history
operator: eq
value_type: bool
value_bool: true
canonical_key: kb-requirements-req-core-attention-item-store-md:core.attention_items:one_append_only_store_with_kind_subject_state_history:eq:true
claim_key: CLAIM-C09B827CC1A61D1B
claim_text: Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:44:41.904Z'
id: FACT-PROP-3126E40F7A919F1C
type: fact
---
