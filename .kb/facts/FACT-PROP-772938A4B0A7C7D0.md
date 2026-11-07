---
title: Base Url Https Only = true
status: active
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.endpoint
property_key: base_url_https_only
canonical_key: mcp-kibi-model-requirement:sdk.endpoint:base_url_https_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7C05C0F7460E7BCF
claim_text: A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:07.491Z'
id: FACT-PROP-772938A4B0A7C7D0
type: fact
---
