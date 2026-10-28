---
title: Operator Cluster Topology
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.cluster_topology
canonical_key: operator.cluster_topology
claim_key: CLAIM-89E9516DE59E88C9
claim_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:37.727Z'
id: FACT-SUBJECT-E4ED656B951F9EFF
type: fact
---
