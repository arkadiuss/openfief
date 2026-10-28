---
title: Matches Openapi Spec = true
status: active
text_ref: docs/DESIGN.md#D-50
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-50
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.read_api
property_key: matches_openapi_spec
canonical_key: mcp-kibi-model-requirement:views.read_api:matches_openapi_spec:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F8D92AA3C4096CAA
claim_text: 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.'
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:40.305Z'
id: FACT-PROP-A0A73C80204CB2B7
type: fact
---
