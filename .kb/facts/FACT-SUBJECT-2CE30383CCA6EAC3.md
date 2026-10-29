---
title: Operator Workload Objects
status: active
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.workload_objects
canonical_key: operator.workload_objects
claim_key: CLAIM-2BB136AFA61FBAE7
claim_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:12:55.409Z'
id: FACT-SUBJECT-2CE30383CCA6EAC3
type: fact
---
