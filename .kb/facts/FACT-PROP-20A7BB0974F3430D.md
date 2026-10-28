---
title: Authenticated Per Call = true
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
subject_key: ingest.core_connector_calls
property_key: authenticated_per_call
canonical_key: mcp-kibi-model-requirement:ingest.core_connector_calls:authenticated_per_call:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4F2CFB6B6AD1A8CB
claim_text: Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:56:10.167Z'
id: FACT-PROP-20A7BB0974F3430D
type: fact
---
