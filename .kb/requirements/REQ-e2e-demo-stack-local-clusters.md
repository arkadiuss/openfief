---
title: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.
status: open
text_ref: docs/DESIGN.md#D-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-12
  - lane:strict
semantic_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.
logic_claims:
  - CLAIM-E57CA9DAB6653EF4
semantic_clauses:
  - The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2ef617925752ac64994439adafca9c2fb8f4da7d51dcbb614ed28dbab25e2a9b
semantic_inventory:
  - claim_key: CLAIM-E57CA9DAB6653EF4
    claim_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route
    role: normative
    status: modeled
    span:
      start: 0
      end: 280
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:16.545Z'
id: REQ-e2e-demo-stack-local-clusters
type: req
---
The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core reachable through its Service from admitted pods only and, on k3s, through the demo's Traefik route.

## Context

D-12 says OpenFief runs on any conformant cluster and is tested on kind and k3d, with a CNI that enforces NetworkPolicy and a storage class as documented requirements. The orchestrating agent asked for an end-to-end suite after the live demo on k3s answered 503 from Traefik because the core pod never became ready and nobody could see why; the author wants CI, not SSH, to be how the demo's stack is debugged, so the k3d cluster runs the demo machine's k3s version with its kube-router policy controller and Traefik. Modeled by Claude Code; the author has not reviewed it.

## Source

> **[D-12]** Any conformant Kubernetes cluster; no reference platform. Tested on kind and k3d, with documented requirements: a CNI that enforces NetworkPolicy, and a storage class.

docs/DESIGN.md#D-12
