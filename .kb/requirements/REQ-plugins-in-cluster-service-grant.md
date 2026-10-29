---
title: 'A plugin''s signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-1
  - lane:strict
text_ref: design-decisions:D-74#1
origin:
  kind: import
  ref: design-decisions:D-74#1
  recorded_at: '2026-10-29T07:35:38.538Z'
semantic_text: 'A plugin''s signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service.'
logic_claims:
  - CLAIM-886737E753EBB6EE
semantic_clauses:
  - 'A plugin''s signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6d4d53bee6fde0b8c2dc741210236d3bcca36b34e083a70d1c08c19ba6c6cdc7
semantic_inventory:
  - claim_key: CLAIM-886737E753EBB6EE
    claim_text: 'A plugin''s signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service'
    role: normative
    status: modeled
    span:
      start: 0
      end: 216
id: REQ-plugins-in-cluster-service-grant
type: req
---
A plugin's signed manifest may request one in-cluster service, which the owner must pick at install: the CLI must ask for its namespace, service and port and show them, and the grant must approve exactly that service.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. D-74 and D-75 answer the implementation's question that the Ghostfolio view could not reach the owner's Ghostfolio: a manifest is the plugin's signed image label (D-42), so an image built and signed in CI can't name a service or host the owner picks at install. D-74 covers a service inside the cluster, such as a Ghostfolio running in the same cluster; D-16 keeps what the manifest requests apart from what the owner's grant approves. D-74 is an exception to D-32, whose rules REQ-AUTO-3C8B1ED0D820F0C9 and REQ-AUTO-8CBFBD6A12E0865B still state without it. The author recorded D-74 with D-73 and D-75 after the design session that answered the implementation's questions.

## Source

> An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed. This traffic doesn't pass through the egress proxy, and it may be plain HTTP, which the install warns about. When the service's namespace has its own NetworkPolicies, the CLI prints the ingress rule to add there.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-74#1
