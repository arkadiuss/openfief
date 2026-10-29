---
title: Reads Operator Environment = true
status: active
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.go.plugin_environment
property_key: reads_operator_environment
canonical_key: mcp-kibi-model-requirement:sdk.go.plugin_environment:reads_operator_environment:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-326E998054C10CA9
claim_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:29.951Z'
id: FACT-PROP-BDD5DBF12CAFE675
type: fact
---
