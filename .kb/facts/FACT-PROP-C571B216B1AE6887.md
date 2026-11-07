---
title: Exposure By Component Only = true
status: active
text_ref: design-decisions:D-200#13
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-13
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: exposure_by_component_only
canonical_key: mcp-kibi-model-requirement:m0.exit:exposure_by_component_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EAB79D0283387F2F
claim_text: The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:22.534Z'
id: FACT-PROP-C571B216B1AE6887
type: fact
---
