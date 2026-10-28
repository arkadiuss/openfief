---
title: Enforces Same Origin And Strict Csp = true
status: active
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.built_in_ui
property_key: enforces_same_origin_and_strict_csp
canonical_key: mcp-kibi-model-requirement:views.built_in_ui:enforces_same_origin_and_strict_csp:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0318328C509EFF88
claim_text: The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:43.585Z'
id: FACT-PROP-DB3A4017DE2B5713
type: fact
---
