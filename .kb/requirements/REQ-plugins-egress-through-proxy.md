---
title: Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-1
  - provenance:design-decisions-d-74
  - lane:strict
text_ref: design-decisions:D-32#1
origin:
  kind: import
  ref: design-decisions:D-32#1
  recorded_at: '2026-10-29T07:52:46.907Z'
semantic_text: Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it.
logic_claims:
  - CLAIM-53790655D40A543B
semantic_clauses:
  - Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 18ac2c3bc5cbcaeff3e9197ce5a1f9beceb576f625979a8175b320f69b940495
semantic_inventory:
  - claim_key: CLAIM-53790655D40A543B
    claim_text: Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it
    role: exception
    status: modeled
    span:
      start: 0
      end: 129
id: REQ-plugins-egress-through-proxy
type: req
---
Plugin egress must go through the egress proxy, except a plugin's traffic to the one in-cluster service the owner approved for it.

## Why this changed

The author's decision D-74 makes an explicit exception to D-32 for services inside the cluster: a plugin may reach one in-cluster service that the owner picks at install and the grant approves exactly, and that traffic doesn't pass through the egress proxy and may be plain HTTP. REQ-AUTO-8CBFBD6A12E0865B stated D-32's rule that plugin egress goes through the egress proxy without that exception, so this requirement replaces it; every other host, including a host taken from a config field (D-75), still goes through the proxy.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames. D-74's exception is recorded in REQ-plugins-in-cluster-service-grant, REQ-plugins-in-cluster-service-no-proxy and REQ-plugins-in-cluster-service-plain-http-warning; REQ-plugins-config-field-host-via-proxy records that D-75's hosts still use the proxy. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library. The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list. Plugins identify themselves with their projected ServiceAccount token, through a custom role function. NetworkPolicy lets plugins reach only the core and the proxy, with no DNS, and the proxy refuses private and internal address ranges. Plugins reach external hosts through `HTTPS_PROXY`.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-32#1

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74
