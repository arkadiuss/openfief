---
title: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, plus the pods of its granted in-cluster service on that service's pod port, and never DNS.
status: open
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
semantic_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, plus the pods of its granted in-cluster service on that service's pod port, and never DNS.
logic_claims:
  - CLAIM-D2B82DD8C82DFA57
semantic_clauses:
  - A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, plus the pods of its granted in-cluster service on that service's pod port, and never DNS.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5a8dfd8f8d5b4e452865fc12943b835f8c3b26f9ed7820d80bcc7d16e6181744
semantic_inventory:
  - claim_key: CLAIM-D2B82DD8C82DFA57
    claim_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, plus the pods of its granted in-cluster service on that service's pod port, and never DNS
    role: normative
    status: modeled
    span:
      start: 0
      end: 260
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:26.996Z'
id: REQ-plugins-network-policy-core-and-proxy
type: req
---
A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, plus the pods of its granted in-cluster service on that service's pod port, and never DNS.

## Context

D-32 limits plugins to the core and the egress proxy without DNS; D-74 adds one in-cluster service the owner approved, reached directly. Refined for D-74 by the plugin lifecycle package.

## Source

> [D-74] The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed.

docs/DESIGN.md#D-74
