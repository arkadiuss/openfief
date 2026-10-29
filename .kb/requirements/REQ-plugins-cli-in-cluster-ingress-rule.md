---
title: When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-5
  - lane:strict
text_ref: design-decisions:D-74#5
origin:
  kind: import
  ref: design-decisions:D-74#5
  recorded_at: '2026-10-29T07:38:30.837Z'
semantic_text: When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there.
logic_claims:
  - CLAIM-C05556EC595A60D7
semantic_clauses:
  - When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8b36d534bff0c3672fde49d8356972e0ada03350f1a0bc47b7546c122ca6bfb6
semantic_inventory:
  - claim_key: CLAIM-C05556EC595A60D7
    claim_text: When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there
    role: condition
    status: modeled
    span:
      start: 0
      end: 125
id: REQ-plugins-cli-in-cluster-ingress-rule
type: req
---
When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. D-74 has the operator add the plugin's egress rule to the one in-cluster service its grant approves (REQ-operator-in-cluster-service-route); NetworkPolicies in the service's own namespace govern what may reach that service's pods. The operator holds permissions only in the openfief namespace (D-15), and the owner applies resources with their own kubeconfig (D-46); the decision states no further reason for printing the rule. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74#5
