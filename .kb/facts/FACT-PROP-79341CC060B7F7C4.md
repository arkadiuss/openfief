---
title: Relays Login Through Auth Service = true
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
property_key: relays_login_through_auth_service
canonical_key: mcp-kibi-model-requirement:views.built_in_ui:relays_login_through_auth_service:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B2CE8B3F09E4F95E
claim_text: 'The built-in UI must sign the owner in only through the core''s AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session.'
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:26.687Z'
id: FACT-PROP-79341CC060B7F7C4
type: fact
---
