---
title: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and the numeric target ports of the ports plugins dial, the core's being its connect port, which targets the core's plugin port, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and the numeric target ports of the ports plugins dial, the core's being its connect port, which targets the core's plugin port, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
logic_claims:
  - CLAIM-51CF57BD5449D2C0
semantic_clauses:
  - The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and the numeric target ports of the ports plugins dial, the core's being its connect port, which targets the core's plugin port, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2e11d9457a7f27765f1843b11d10bb181d808358d62e176f170973f41c2b9782
semantic_inventory:
  - claim_key: CLAIM-51CF57BD5449D2C0
    claim_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and the numeric target ports of the ports plugins dial, the core's being its connect port, which targets the core's plugin port, and must refuse a Service without a pod selector or whose selector also matches plugin pods
    role: normative
    status: modeled
    span:
      start: 0
      end: 315
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:42.954Z'
id: REQ-operator-cluster-peers
type: req
---
The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and the numeric target ports of the ports plugins dial, the core's being its connect port, which targets the core's plugin port, and must refuse a Service without a pod selector or whose selector also matches plugin pods.

## Context

Plugins have no DNS (D-32), so the operator hands them ClusterIPs and ports, and NetworkPolicy matches pod ports, so the target ports must be numbers. D-115 gives the core a plugin port behind the Service's connect port, which the operator's --core-port names, so every plugin's core address and egress rule name that port and never the browser port.

## Source

> The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79).

docs/DESIGN.md#D-115
