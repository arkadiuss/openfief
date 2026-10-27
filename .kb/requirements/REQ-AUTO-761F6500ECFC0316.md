---
title: Each plugin must ship a manifest that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC.
status: open
text_ref: design-decisions:D-11#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-11-3
  - lane:strict
semantic_text: Each plugin must ship a manifest that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC.
logic_claims:
  - CLAIM-80846517A5A8605B
semantic_clauses:
  - Each plugin must ship a manifest that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9ac2289c7edef27860b18e709ac1971635481e01ad7d21301d69ad0776d16ce9
semantic_inventory:
  - claim_key: CLAIM-80846517A5A8605B
    claim_text: Each plugin must ship a manifest that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC
    role: normative
    status: modeled
    span:
      start: 0
      end: 98
origin:
  kind: import
  ref: design-decisions:D-11#3
  recorded_at: '2026-10-27T20:12:29.127Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-761F6500ECFC0316
type: req
---
Each plugin must ship a manifest that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC.

## Source

> The core is written in Go. Each plugin runs in its own container and speaks protobuf contracts over Connect RPC. Each plugin ships a manifest (in the style of OpenClaw's, R-12) that OpenFief compiles into NetworkPolicy, Secret mounts and RBAC. SDKs come in TypeScript and Go. Egress hostnames are enforced by the egress proxy (D-32), because NetworkPolicy has no hostnames.

Source: OpenFief author decisions (DESIGN.md section 10) - D-11#3
