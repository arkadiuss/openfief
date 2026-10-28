---
title: Separates Plugin And Session Tokens = true
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.built_in_ui
property_key: separates_plugin_and_session_tokens
canonical_key: mcp-kibi-model-requirement:views.built_in_ui:separates_plugin_and_session_tokens:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3581F50D38DBDE86
claim_text: The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:32.333Z'
id: FACT-PROP-805C05CD3415A9DC
type: fact
---
