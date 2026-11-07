---
title: Imports Idempotent And Versioned = true
status: active
text_ref: design-decisions:D-200#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: imports_idempotent_and_versioned
canonical_key: mcp-kibi-model-requirement:m0.exit:imports_idempotent_and_versioned:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-16A8D2D3F13F3B32
claim_text: Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:34.209Z'
id: FACT-PROP-F8C0B13CB66E10E0
type: fact
---
