---
title: Limits Plugin To Core And Proxy = true
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
subject_key: plugins.network_policy
property_key: limits_plugin_to_core_and_proxy
canonical_key: mcp-kibi-model-requirement:plugins.network_policy:limits_plugin_to_core_and_proxy:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A8DCAEEA129FE717
claim_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:25.044Z'
id: FACT-PROP-293DE0CDA2A958C8
type: fact
---
