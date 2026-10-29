---
title: Goes Through Egress Proxy Except Approved In Cluster Service = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-32#1
fact_kind: property_value
subject_key: plugins.plugin_egress
property_key: goes_through_egress_proxy_except_approved_in_cluster_service
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:plugins.plugin_egress:goes_through_egress_proxy_except_approved_in_cluster_service:eq:true
claim_key: CLAIM-53790655D40A543B
claim_text: Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:52:32.048Z'
id: FACT-PROP-3217D09BE0B8804E
type: fact
---
