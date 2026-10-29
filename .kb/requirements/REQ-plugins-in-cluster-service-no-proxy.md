---
title: A plugin's traffic to its granted in-cluster service must not pass through the egress proxy.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-3
  - lane:strict
text_ref: design-decisions:D-74#3
origin:
  kind: import
  ref: design-decisions:D-74#3
  recorded_at: '2026-10-29T07:37:08.767Z'
semantic_text: A plugin's traffic to its granted in-cluster service must not pass through the egress proxy.
logic_claims:
  - CLAIM-738E1C01C990C32C
semantic_clauses:
  - A plugin's traffic to its granted in-cluster service must not pass through the egress proxy.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e928d2a68e401d71cd4bbde307199259d50787f2f9faf7a1be6ae9edf2e1dc84
semantic_inventory:
  - claim_key: CLAIM-738E1C01C990C32C
    claim_text: A plugin's traffic to its granted in-cluster service must not pass through the egress proxy
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
id: REQ-plugins-in-cluster-service-no-proxy
type: req
---
A plugin's traffic to its granted in-cluster service must not pass through the egress proxy.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-74 lets a plugin reach the one in-cluster service its grant approves (REQ-plugins-in-cluster-service-grant) over a NetworkPolicy egress rule to that service's pods on the granted port and the service's ClusterIP (REQ-operator-in-cluster-service-route), and that traffic goes around the egress proxy; the decision states no further reason. D-74 is an exception to D-32: REQ-AUTO-8CBFBD6A12E0865B still states D-32's rule that plugin egress goes through the egress proxy, and REQ-AUTO-3C8B1ED0D820F0C9 its rule that plugins reach only the core and the proxy. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74#3
