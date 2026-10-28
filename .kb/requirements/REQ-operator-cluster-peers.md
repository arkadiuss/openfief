---
title: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
logic_claims:
  - CLAIM-89E9516DE59E88C9
semantic_clauses:
  - The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 61d91fecb976891968e234e9c9b7575aafbe067d4bbe52bef50123b37898358b
semantic_inventory:
  - claim_key: CLAIM-89E9516DE59E88C9
    claim_text: The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods
    role: normative
    status: modeled
    span:
      start: 0
      end: 212
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:42.954Z'
id: REQ-operator-cluster-peers
type: req
---
The operator must take the core's and egress proxy's addresses from their Services' ClusterIPs and numeric target ports, and must refuse a Service without a pod selector or whose selector also matches plugin pods.

## Context

D-32 gives plugins no DNS, so they reach the core and the proxy by ClusterIP, and their NetworkPolicy names the core and the proxy as the only peers. An empty selector would select every pod, and a selector that matches plugin pods would let one plugin pass as the core or the proxy, so both are refused and no plugin runs. Refines REQ-AUTO-3C8B1ED0D820F0C9. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> NetworkPolicy lets plugins reach only the core and the proxy, with no DNS

docs/DESIGN.md D-32
