---
title: No Fixed Token Or Bypass = true
status: active
text_ref: docs/DESIGN.md#D-84
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-84
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.authentication
property_key: no_fixed_token_or_bypass
canonical_key: mcp-kibi-model-requirement:identity.authentication:no_fixed_token_or_bypass:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-470CEEBB591CE491
claim_text: 'No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token.'
origin:
  kind: agent
  recorded_at: '2026-10-29T10:54:20.976Z'
id: FACT-PROP-D1CEBAEA762603BB
type: fact
---
