---
title: Routes Only Granted Egress Through Proxy = true
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
subject_key: plugins.egress_route
property_key: routes_only_granted_egress_through_proxy
canonical_key: mcp-kibi-model-requirement:plugins.egress_route:routes_only_granted_egress_through_proxy:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B8823E04FCA2649D
claim_text: A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:39.306Z'
id: FACT-PROP-A6E61A86C4B8338E
type: fact
---
