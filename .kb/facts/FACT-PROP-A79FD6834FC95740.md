---
title: Uses Vetted Webauthn And Refuses Clones = true
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
subject_key: identity.passkey
property_key: uses_vetted_webauthn_and_refuses_clones
canonical_key: mcp-kibi-model-requirement:identity.passkey:uses_vetted_webauthn_and_refuses_clones:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F6D6C01F4DA9E636
claim_text: Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:25.507Z'
id: FACT-PROP-A79FD6834FC95740
type: fact
---
