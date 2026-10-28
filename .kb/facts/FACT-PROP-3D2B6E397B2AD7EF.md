---
title: Derive Peers From Services = true
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
subject_key: operator.cluster_topology
property_key: derive_peers_from_services
canonical_key: mcp-kibi-model-requirement:operator.cluster_topology:derive_peers_from_services:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-89E9516DE59E88C9
claim_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:40.354Z'
id: FACT-PROP-3D2B6E397B2AD7EF
type: fact
---
