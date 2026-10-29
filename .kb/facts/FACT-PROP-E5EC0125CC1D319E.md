---
title: Instance Scoped Short Lived = true
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
subject_key: identity.core_token
property_key: instance_scoped_short_lived
canonical_key: mcp-kibi-model-requirement:identity.core_token:instance_scoped_short_lived:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5C8E0FE7BAC6701B
claim_text: The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:45.692Z'
id: FACT-PROP-E5EC0125CC1D319E
type: fact
---
