---
title: Accepts View Plugin Tokens = true
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
subject_key: identity.read_api_auth
property_key: accepts_view_plugin_tokens
canonical_key: mcp-kibi-model-requirement:identity.read_api_auth:accepts_view_plugin_tokens:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EA4B749449BAFB02
claim_text: The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:17.980Z'
id: FACT-PROP-95AA4D2FBA48DB6B
type: fact
---
