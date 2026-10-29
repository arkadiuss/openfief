---
title: A plugin's traffic to its granted in-cluster service may be plain HTTP, and the install must warn the owner about that.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-4
  - lane:strict
text_ref: design-decisions:D-74#4
origin:
  kind: import
  ref: design-decisions:D-74#4
  recorded_at: '2026-10-29T07:37:49.140Z'
semantic_text: A plugin's traffic to its granted in-cluster service may be plain HTTP, and the install must warn the owner about that.
logic_claims:
  - CLAIM-E565A46170840634
semantic_clauses:
  - A plugin's traffic to its granted in-cluster service may be plain HTTP, and the install must warn the owner about that.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 421c2bafd4799477afb9a138a9d9d111ea6b0de701c1a36f0ee5faa8e53b7aa5
semantic_inventory:
  - claim_key: CLAIM-E565A46170840634
    claim_text: A plugin's traffic to its granted in-cluster service may be plain HTTP, and the install must warn the owner about that
    role: normative
    status: modeled
    span:
      start: 0
      end: 118
id: REQ-plugins-in-cluster-service-plain-http-warning
type: req
---
A plugin's traffic to its granted in-cluster service may be plain HTTP, and the install must warn the owner about that.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. D-32 has plugins reach external hosts through HTTPS_PROXY; D-74 lets a plugin reach the one in-cluster service its grant approves (REQ-plugins-in-cluster-service-grant) around the egress proxy (REQ-plugins-in-cluster-service-no-proxy), such as a Ghostfolio running in the same cluster, and allows that traffic to be plain HTTP provided the install warns about it; the decision states no further reason. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74#4
