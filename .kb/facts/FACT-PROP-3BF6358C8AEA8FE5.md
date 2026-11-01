---
title: Relayed In Order = true
status: active
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.outbox
property_key: relayed_in_order
canonical_key: mcp-kibi-model-requirement:attention.outbox:relayed_in_order:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9FF3D9D9E3634B7A
claim_text: The outbox's events must be published in order, at least once, resuming at the first event a publisher failed on, and the log publisher must not log an item's subject or detail.
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:05.408Z'
id: FACT-PROP-3BF6358C8AEA8FE5
type: fact
---
