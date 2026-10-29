---
title: For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-2
  - lane:strict
text_ref: design-decisions:D-74#2
origin:
  kind: import
  ref: design-decisions:D-74#2
  recorded_at: '2026-10-29T07:36:26.116Z'
semantic_text: For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS.
logic_claims:
  - CLAIM-0006282949A36DDA
semantic_clauses:
  - For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 136b522dad17929b9bc0e031b017a268e03b7d60e54bc9b184fdb617b1d1b9d9
semantic_inventory:
  - claim_key: CLAIM-0006282949A36DDA
    claim_text: For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS
    role: normative
    status: modeled
    span:
      start: 0
      end: 229
id: REQ-operator-in-cluster-service-route
type: req
---
For a plugin's granted in-cluster service, the operator must add a NetworkPolicy egress rule from the plugin to that service's pods on that port and give the plugin the service's ClusterIP address, so that the plugin needs no DNS.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. D-74 lets a plugin reach the one in-cluster service its grant approves (REQ-plugins-in-cluster-service-grant). Plugins have no DNS (D-32), so the operator already gives them the core's and the egress proxy's ClusterIPs (REQ-operator-cluster-peers); D-74 gives them the granted service's ClusterIP the same way. D-74 is an exception to D-32: REQ-AUTO-3C8B1ED0D820F0C9 still states D-32's rule that NetworkPolicy lets plugins reach only the core and the proxy, as do the implementation's refinements REQ-plugins-network-policy-core-and-proxy and REQ-operator-workload-objects. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74#2
