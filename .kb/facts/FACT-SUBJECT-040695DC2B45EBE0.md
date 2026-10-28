---
title: Operator Plugin Objects
status: active
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.plugin_objects
canonical_key: operator.plugin_objects
claim_key: CLAIM-58700B0AF172714D
claim_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:29.247Z'
id: FACT-SUBJECT-040695DC2B45EBE0
type: fact
---
