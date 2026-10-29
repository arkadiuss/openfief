---
title: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service.
status: open
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
semantic_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service.
logic_claims:
  - CLAIM-2BB136AFA61FBAE7
semantic_clauses:
  - The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6ffce2a48197d160cd42fb3402430bbb7fd2a694088a258e80805a657f86eb57
semantic_inventory:
  - claim_key: CLAIM-2BB136AFA61FBAE7
    claim_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service
    role: exception
    status: modeled
    span:
      start: 0
      end: 337
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:34.245Z'
id: REQ-operator-workload-objects
type: req
---
The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods and to the pods of its granted in-cluster service.

## Context

D-11 compiles manifests into workloads; D-74 adds the rule to an approved in-cluster service's pods. Refined for D-74 by the plugin lifecycle package.

## Source

> [D-74] The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port.

docs/DESIGN.md#D-74
