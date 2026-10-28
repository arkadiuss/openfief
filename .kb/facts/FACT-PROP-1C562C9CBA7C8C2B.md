---
title: Fails Closed = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_allow_list
property_key: fails_closed
canonical_key: mcp-kibi-model-requirement:plugins.egress_allow_list:fails_closed:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FF247BD182B6C88B
claim_text: The egress proxy must follow the allow-list ConfigMap as it changes and, while it is missing or cannot be decoded, treat every plugin as having no egress allow-list.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:36.470Z'
id: FACT-PROP-1C562C9CBA7C8C2B
type: fact
---
