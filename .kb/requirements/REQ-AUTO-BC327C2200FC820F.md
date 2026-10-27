---
title: Plugin-to-core communication must use protobuf contracts over Connect RPC.
status: open
text_ref: design-decisions:D-11#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-11-2
  - lane:strict
semantic_text: Plugin-to-core communication must use protobuf contracts over Connect RPC.
logic_claims:
  - CLAIM-8C7DC82B87A66D33
semantic_clauses:
  - Plugin-to-core communication must use protobuf contracts over Connect RPC.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2f2b0c8cde618cc68530c24b25a4c6187b030163f811b8c48b2c0d11db1f010e
semantic_inventory:
  - claim_key: CLAIM-8C7DC82B87A66D33
    claim_text: Plugin-to-core communication must use protobuf contracts over Connect RPC
    role: normative
    status: modeled
    span:
      start: 0
      end: 73
origin:
  kind: import
  ref: design-decisions:D-11#2
  recorded_at: '2026-10-27T20:12:28.901Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-BC327C2200FC820F
type: req
---
Plugin-to-core communication must use protobuf contracts over Connect RPC.

## Source

> The core is written in Go. Each plugin runs in its own container and speaks protobuf contracts over Connect RPC. Each plugin ships a manifest (in the style of OpenClaw's, R-12) that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC. SDKs come in TypeScript and Go. Egress hostnames are enforced by the egress proxy (D-32), because NetworkPolicy has no hostnames.

Source: OpenFief author decisions (DESIGN.md section 10) - D-11#2
