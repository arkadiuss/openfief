---
title: Checks Every Resolved Address = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: checks_every_resolved_address
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:checks_every_resolved_address:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D99BF274B0E8BB3F
claim_text: The egress proxy must resolve an allowed hostname itself and refuse the request when any resolved address fails the forbidden-address check, dialing only checked addresses.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:28.857Z'
id: FACT-PROP-054A8D23C60FEA2E
type: fact
---
