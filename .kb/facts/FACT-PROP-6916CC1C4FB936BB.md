---
title: Keeps Api Key Out Of Errors = true
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.eodhd_adapter
property_key: keeps_api_key_out_of_errors
canonical_key: mcp-kibi-model-requirement:valuation.eodhd_adapter:keeps_api_key_out_of_errors:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8C920A6BB062F2BB
claim_text: The EODHD adapter must send the owner's API key only as EODHD's api_token parameter, through the HTTP client and context its caller supplies, must never return the key in an error, must treat 404 as an unknown ticker and must fail on any other status than 200 or 404.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:34.620Z'
id: FACT-PROP-6916CC1C4FB936BB
type: fact
---
