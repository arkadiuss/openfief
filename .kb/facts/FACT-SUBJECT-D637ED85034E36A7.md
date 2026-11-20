---
title: Plugins In Cluster Service
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-74#1
fact_kind: subject
subject_key: plugins.in_cluster_service
canonical_key: plugins.in_cluster_service
claim_key: CLAIM-886737E753EBB6EE
claim_text: 'A plugin''s signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service.'
origin:
  kind: agent
  recorded_at: '2026-10-29T07:35:23.247Z'
id: FACT-SUBJECT-D637ED85034E36A7
type: fact
---

Definition not stated: the subject plugins.in_cluster_service is named by requirements REQ-operator-in-cluster-service-route, REQ-plugins-cli-in-cluster-ingress-rule, REQ-plugins-in-cluster-service-grant, REQ-plugins-in-cluster-service-no-proxy, REQ-plugins-in-cluster-service-plain-http-warning, REQ-plugins-in-cluster-service-read-role, REQ-plugins-in-cluster-service-role-removed, which link to the subject fact FACT-SUBJECT-D637ED85034E36A7 through constrains. The people who own that source can say what it is in the product and what it is not.
