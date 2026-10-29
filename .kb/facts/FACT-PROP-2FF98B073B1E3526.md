---
title: Traffic Passes Through Egress Proxy = false
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-3
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-74#3
fact_kind: property_value
subject_key: plugins.in_cluster_service
property_key: traffic_passes_through_egress_proxy
operator: eq
value_type: bool
value_bool: false
canonical_key: docs-design-md:plugins.in_cluster_service:traffic_passes_through_egress_proxy:eq:false
claim_key: CLAIM-738E1C01C990C32C
claim_text: A plugin's traffic to its granted in-cluster service must not pass through the egress proxy.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:36:58.156Z'
id: FACT-PROP-2FF98B073B1E3526
type: fact
---
