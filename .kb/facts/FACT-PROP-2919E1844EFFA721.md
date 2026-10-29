---
title: Operator Adds Egress Rule And Cluster Ip = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-74#2
fact_kind: property_value
subject_key: plugins.in_cluster_service
property_key: operator_adds_egress_rule_and_cluster_ip
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:plugins.in_cluster_service:operator_adds_egress_rule_and_cluster_ip:eq:true
claim_key: CLAIM-0006282949A36DDA
claim_text: For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:36:14.378Z'
id: FACT-PROP-2919E1844EFFA721
type: fact
---
