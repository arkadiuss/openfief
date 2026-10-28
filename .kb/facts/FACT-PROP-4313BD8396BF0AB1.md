---
title: Admits Only The Core = true
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.go.connector_server
property_key: admits_only_the_core
canonical_key: mcp-kibi-model-requirement:sdk.go.connector_server:admits_only_the_core:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-737BEEAD3377E06B
claim_text: The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:24.674Z'
id: FACT-PROP-4313BD8396BF0AB1
type: fact
---
