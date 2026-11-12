---
title: Walks Read Api Contract = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: e2e.apicontract
property_key: walks_read_api_contract
canonical_key: mcp-kibi-model-requirement:e2e.apicontract:walks_read_api_contract:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-05B4C19D1D4FAB21
claim_text: The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call.
origin:
  kind: agent
  recorded_at: '2026-11-08T04:07:09.991Z'
id: FACT-PROP-E3B75F0406656DEC
type: fact
---
