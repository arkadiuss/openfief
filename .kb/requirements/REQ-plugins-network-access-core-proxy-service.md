---
title: Plugin network access must be limited to the core, the egress proxy and at most one in-cluster service the owner approved for that plugin, with no DNS.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-4
  - provenance:design-decisions-d-74
  - lane:strict
text_ref: design-decisions:D-32#4
origin:
  kind: import
  ref: design-decisions:D-32#4
  recorded_at: '2026-10-29T07:51:49.535Z'
semantic_text: Plugin network access must be limited to the core, the egress proxy and at most one in-cluster service the owner approved for that plugin, with no DNS.
logic_claims:
  - CLAIM-C2DDE504C91EA626
semantic_clauses:
  - Plugin network access must be limited to the core, the egress proxy and at most one in-cluster service the owner approved for that plugin, with no DNS.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1df2923ac0ab16734c230b8f8fdde6efbd0c635a9e1bab745867aea54c5d0036
semantic_inventory:
  - claim_key: CLAIM-C2DDE504C91EA626
    claim_text: Plugin network access must be limited to the core, the egress proxy and at most one in-cluster service the owner approved for that plugin, with no DNS
    role: normative
    status: modeled
    span:
      start: 0
      end: 150
id: REQ-plugins-network-access-core-proxy-service
type: req
---
Plugin network access must be limited to the core, the egress proxy and at most one in-cluster service the owner approved for that plugin, with no DNS.

## Why this changed

The author's decision D-74 makes an explicit exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, the grant approves exactly that service, and the operator adds a NetworkPolicy egress rule from the plugin to that service's pods on that port. REQ-AUTO-3C8B1ED0D820F0C9 stated D-32's rule that plugins reach only the core and the proxy without that exception, so this requirement replaces it and keeps the rest of D-32's rule, including no DNS.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-74's exception is recorded in REQ-plugins-in-cluster-service-grant and REQ-operator-in-cluster-service-route. The implementation's refinements REQ-plugins-network-policy-core-and-proxy and REQ-operator-workload-objects still describe the rule without the exception. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library. The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list. Plugins identify themselves with their projected ServiceAccount token, through a custom role function. NetworkPolicy lets plugins reach only the core and the proxy, with no DNS, and the proxy refuses private and internal address ranges. Plugins reach external hosts through `HTTPS_PROXY`.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-32#4

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74
