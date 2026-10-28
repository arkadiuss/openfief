---
title: Bounded Request With Injected Client = true
status: active
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.nbp_table_a
property_key: bounded_request_with_injected_client
canonical_key: mcp-kibi-model-requirement:valuation.nbp_table_a:bounded_request_with_injected_client:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9F4B9CC015776EA9
claim_text: The NBP adapter must request table A for at most 93 days per call, through the HTTP client and context its caller supplies and with its own User-Agent, and must fail on any status other than 200 or 404.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:57:23.868Z'
id: FACT-PROP-41ADB9320F97CB86
type: fact
---
