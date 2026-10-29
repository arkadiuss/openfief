---
title: The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute.
status: open
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
semantic_text: The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute.
logic_claims:
  - CLAIM-0983A2ABC9FB9A97
semantic_clauses:
  - The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a8aaa7bfe55182ba36304c4e8989758ee2ac9adb16f2d1a08b6e2387e4054ea4
semantic_inventory:
  - claim_key: CLAIM-0983A2ABC9FB9A97
    claim_text: The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute
    role: normative
    status: modeled
    span:
      start: 0
      end: 181
origin:
  kind: agent
  recorded_at: '2026-10-29T15:11:14.386Z'
id: REQ-operator-in-cluster-service-reads-service
type: req
---
The operator must read a plugin's granted in-cluster service outside its namespaced cache, run nothing while that Service is missing or unusable, and reconcile again within a minute.

## Context

D-74 has the operator add a rule to the service's pods and pass its ClusterIP, which needs the Service's selector, ClusterIP and pod port. The operator's cache sees only openfief (D-15) and it doesn't watch other namespaces, so it reads uncached and retries. Refines REQ-operator-in-cluster-service-route.

## Source

> [D-74] An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed.

docs/DESIGN.md#D-74
