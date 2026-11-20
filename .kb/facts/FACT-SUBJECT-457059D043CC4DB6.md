---
title: Plugins Cluster Service
status: archived
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: plugins.cluster_service
canonical_key: plugins.cluster_service
claim_key: CLAIM-E79B8B2B5CC0C084
claim_text: An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:59.556Z'
id: FACT-SUBJECT-457059D043CC4DB6
type: fact
part_of:
  - plugins.in_cluster_service
---
The form of an in-cluster service a Plugin may name, namespace/service:port with DNS-label names and a namespace other than openfief. It is a part of plugins.in_cluster_service, which covers how the owner picks and approves it.