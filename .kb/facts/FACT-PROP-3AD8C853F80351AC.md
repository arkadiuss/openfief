---
title: Listed Read Only = true
status: active
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.installed_plugins
property_key: listed_read_only
canonical_key: mcp-kibi-model-requirement:views.installed_plugins:listed_read_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-611003F4F66D1654
claim_text: The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them.
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:43.896Z'
id: FACT-PROP-3AD8C853F80351AC
type: fact
---
