---
title: Serves On Kind And K3d With Netpol = true
status: active
text_ref: docs/DESIGN.md#D-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: deploy.demo_stack
property_key: serves_on_kind_and_k3d_with_netpol
canonical_key: mcp-kibi-model-requirement:deploy.demo_stack:serves_on_kind_and_k3d_with_netpol:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E57CA9DAB6653EF4
claim_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:13.627Z'
id: FACT-PROP-426E35DA6AEADB48
type: fact
---
