---
title: Operator In Cluster Service
status: active
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: operator.in_cluster_service
canonical_key: operator.in_cluster_service
claim_key: CLAIM-0983A2ABC9FB9A97
claim_text: The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:11:09.092Z'
id: FACT-SUBJECT-C9C7F3BE6B6C8D05
type: fact
---

Definition not stated: the subject operator.in_cluster_service is named by requirement REQ-operator-in-cluster-service-reads-service, which links to the subject fact FACT-SUBJECT-C9C7F3BE6B6C8D05 through constrains. The people who own that source can say what it is in the product and what it is not.
