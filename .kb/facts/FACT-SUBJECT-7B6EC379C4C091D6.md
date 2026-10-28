---
title: Deploy Demo Stack
status: active
text_ref: docs/DESIGN.md#D-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-12
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: deploy.demo_stack
canonical_key: deploy.demo_stack
claim_key: CLAIM-E57CA9DAB6653EF4
claim_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:11.099Z'
id: FACT-SUBJECT-7B6EC379C4C091D6
type: fact
---
