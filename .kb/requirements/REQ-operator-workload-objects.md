---
title: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.
status: open
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
semantic_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.
logic_claims:
  - CLAIM-58700B0AF172714D
semantic_clauses:
  - The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c88cbddc130588002e575b7f20f7f2145402f55f358c8efabb6b4caeb8a364c0
semantic_inventory:
  - claim_key: CLAIM-58700B0AF172714D
    claim_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods
    role: exception
    status: modeled
    span:
      start: 0
      end: 287
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:34.245Z'
id: REQ-operator-workload-objects
type: req
---
The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.

## Context

D-11 compiles each manifest into NetworkPolicy, Secret mounts, RBAC and allow-list entries; D-32 limits plugins to the core and the proxy with no DNS; D-27 has plugins present projected tokens; D-29 makes credentials Secrets the owner creates. The operator never creates Secret data, mounts Secrets by name only, runs containers non-root with a read-only root filesystem, and peers are selected by the core's and proxy's Service selectors on pod ports, because NetworkPolicy matches pods after Service translation. Refines REQ-AUTO-761F6500ECFC0316 and REQ-AUTO-3C8B1ED0D820F0C9. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> Each plugin ships a manifest that OpenFief compiles into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries (D-32).

docs/DESIGN.md D-11
