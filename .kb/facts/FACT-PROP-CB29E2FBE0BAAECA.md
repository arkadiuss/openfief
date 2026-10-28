---
title: Keeps Session In Host Only Cookie = true
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
property_key: keeps_session_in_host_only_cookie
canonical_key: mcp-kibi-model-requirement:views.built_in_ui:keeps_session_in_host_only_cookie:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A7102A06A9E2D08D
claim_text: The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:37.988Z'
id: FACT-PROP-CB29E2FBE0BAAECA
type: fact
---
