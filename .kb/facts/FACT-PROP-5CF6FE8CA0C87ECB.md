---
title: Denies Node And Cluster Ranges = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: denies_node_and_cluster_ranges
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:denies_node_and_cluster_ranges:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A6920D0DC066FD88
claim_text: The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse.
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:51.990Z'
id: FACT-PROP-5CF6FE8CA0C87ECB
type: fact
---
