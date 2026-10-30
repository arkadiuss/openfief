---
title: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core's browser port reachable through its Service from admitted pods only, its plugin port from plugin pods only, and, on k3s, the browser port through the demo's Traefik route.
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core's browser port reachable through its Service from admitted pods only, its plugin port from plugin pods only, and, on k3s, the browser port through the demo's Traefik route.
logic_claims:
  - CLAIM-6A92D67E5E3FB186
semantic_clauses:
  - The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core's browser port reachable through its Service from admitted pods only, its plugin port from plugin pods only, and, on k3s, the browser port through the demo's Traefik route.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7af7a3db031af7601c2ab751727c9c21e61e3ee418a10f72f207ac3087d24e9f
semantic_inventory:
  - claim_key: CLAIM-6A92D67E5E3FB186
    claim_text: The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core's browser port reachable through its Service from admitted pods only, its plugin port from plugin pods only, and, on k3s, the browser port through the demo's Traefik route
    role: normative
    status: modeled
    span:
      start: 0
      end: 352
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:16.545Z'
id: REQ-e2e-demo-stack-local-clusters
type: req
---
The demo's stack, installed from the repository's manifests and this checkout's images, must start and serve on both kind and k3d clusters that enforce NetworkPolicy, with the core's browser port reachable through its Service from admitted pods only, its plugin port from plugin pods only, and, on k3s, the browser port through the demo's Traefik route.

## Context

D-12 tests the full flow on a local cluster. The end-to-end suite checks the demo's own policies: its client pod stands where Traefik stands, admitted to the core's http port and refused on the connect port (D-115); the built-in UI, a plugin pod, signs in through the plugin port.

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

docs/DESIGN.md#D-115
