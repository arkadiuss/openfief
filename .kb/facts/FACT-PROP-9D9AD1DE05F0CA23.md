---
title: Requires Verified Bearer Token = true
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
subject_key: views.read_api
property_key: requires_verified_bearer_token
canonical_key: mcp-kibi-model-requirement:views.read_api:requires_verified_bearer_token:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AE6725896A2F811B
claim_text: Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:44:11.244Z'
id: FACT-PROP-9D9AD1DE05F0CA23
type: fact
---
