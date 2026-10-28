---
title: Accepts Only Core Token = true
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
subject_key: sdk.core_verifier
property_key: accepts_only_core_token
canonical_key: mcp-kibi-model-requirement:sdk.core_verifier:accepts_only_core_token:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D1AEC26052D663F0
claim_text: The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:33.930Z'
id: FACT-PROP-947CEFBB39392258
type: fact
---
