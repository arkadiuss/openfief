---
title: An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them.
status: open
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
semantic_text: An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them.
logic_claims:
  - CLAIM-3AA9565159341E1B
semantic_clauses:
  - An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e58ab47e60291a0f9e559a4c7b5344ea9d78ecaab9af2a1d96691a8a428c0f6
semantic_inventory:
  - claim_key: CLAIM-3AA9565159341E1B
    claim_text: An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
origin:
  kind: agent
  recorded_at: '2026-10-29T15:11:21.682Z'
id: REQ-plugins-cli-in-cluster-service-access
type: req
---
An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them.

## Context

D-74 needs the operator to read a Service in another namespace while D-15 keeps its own Role in openfief and D-82 limits cluster-wide permissions. The narrowest reading is a per-instance Role the owner's install creates; recorded as an open question in docs/implementation-questions.md. Refines REQ-plugins-in-cluster-service-grant.

## Source

> [D-74] An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed.

docs/DESIGN.md#D-74
