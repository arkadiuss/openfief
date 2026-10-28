---
title: The operator must not deploy a plugin version whose config fails its manifest's JSON Schema, and must not let a config schema load any other document.
status: open
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
semantic_text: The operator must not deploy a plugin version whose config fails its manifest's JSON Schema, and must not let a config schema load any other document.
logic_claims:
  - CLAIM-7190C277DB20B266
semantic_clauses:
  - The operator must not deploy a plugin version whose config fails its manifest's JSON Schema, and must not let a config schema load any other document.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6b2bc75b354ca336824807cf21bf966691fcdcadd4281ee3b2847906693eefe8
semantic_inventory:
  - claim_key: CLAIM-7190C277DB20B266
    claim_text: The operator must not deploy a plugin version whose config fails its manifest's JSON Schema, and must not let a config schema load any other document
    role: normative
    status: modeled
    span:
      start: 0
      end: 149
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:50.767Z'
id: REQ-operator-config-schema
type: req
---
The operator must not deploy a plugin version whose config fails its manifest's JSON Schema, and must not let a config schema load any other document.

## Context

D-11 has each manifest declare a config schema. Checking the config before deployment keeps a misconfigured version from starting; refusing external references keeps a vendor's schema from making the operator read files or reach the network, which D-15's namespace-only operator has no reason to do. A missing config is checked as an empty object. Refines REQ-AUTO-761F6500ECFC0316. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> Each plugin ships a manifest that OpenFief compiles into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries (D-32).

docs/DESIGN.md D-11
