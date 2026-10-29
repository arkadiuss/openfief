---
title: An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one.
status: open
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
semantic_text: An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one.
logic_claims:
  - CLAIM-E79B8B2B5CC0C084
semantic_clauses:
  - An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ee21a7746921d0bad82bcd7cf4e5cd74a28658637b6611f0cbe87dd1d007250e
semantic_inventory:
  - claim_key: CLAIM-E79B8B2B5CC0C084
    claim_text: An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one
    role: normative
    status: modeled
    span:
      start: 0
      end: 172
origin:
  kind: agent
  recorded_at: '2026-10-29T15:11:04.988Z'
id: REQ-plugins-cluster-service-outside-openfief
type: req
---
An in-cluster service must be namespace/service:port with DNS-label names and a namespace other than openfief, and a Plugin may name one only when its manifest asks for one.

## Context

D-74 allows one in-cluster service the manifest asks for and the owner picks. Refusing the openfief namespace keeps a plugin from using the exception to reach the core or another plugin (D-59); reason for the openfief rule not stated by the author, it is this package's reading. Refines REQ-plugins-in-cluster-service-grant.

## Source

> [D-74] An exception to D-32 for services inside the cluster: a plugin's signed manifest may request one in-cluster service that the owner picks at install, such as a Ghostfolio running in the same cluster. The CLI asks for its namespace, service and port and shows them, and the grant approves exactly that service. The operator then adds a NetworkPolicy egress rule from the plugin to that service's pods on that port and gives the plugin the service's ClusterIP address, so no DNS is needed.

docs/DESIGN.md#D-74
