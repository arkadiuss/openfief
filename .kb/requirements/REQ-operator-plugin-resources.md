---
title: Every plugin container the operator creates must request 50m CPU and 64Mi of memory and be limited to 500m CPU and 256Mi of memory.
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: Every plugin container the operator creates must request 50m CPU and 64Mi of memory and be limited to 500m CPU and 256Mi of memory.
logic_claims:
  - CLAIM-95E1659F18BBD705
semantic_clauses:
  - Every plugin container the operator creates must request 50m CPU and 64Mi of memory and be limited to 500m CPU and 256Mi of memory.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e888da49d69932a0135542ff3699ab8eface202a3b3dbdba802013e92b4d5250
semantic_inventory:
  - claim_key: CLAIM-95E1659F18BBD705
    claim_text: Every plugin container the operator creates must request 50m CPU and 64Mi of memory and be limited to 500m CPU and 256Mi of memory
    role: normative
    status: modeled
    span:
      start: 0
      end: 130
origin:
  kind: agent
  recorded_at: '2026-11-02T20:55:42.829Z'
id: REQ-operator-plugin-resources
type: req
---
Every plugin container the operator creates must request 50m CPU and 64Mi of memory and be limited to 500m CPU and 256Mi of memory.

## Context

The security review found that the operator's plugin pods set no requests or limits and that the install had no LimitRange, so on OpenFief's single small node one plugin's memory leak could take down the node and the PostgreSQL next to it. The author decided that plugin pods get requests and limits (D-182). The values fit the first-party connectors and the built-in UI view with room to spare: a connector syncs on a schedule and a view serves one owner. A manifest-declared size is not built. Refines REQ-operator-workload-objects (GR-12). Modeled by Claude Code for the P0 security fixes before the alpha release, from the security review's finding that plugin pods had no resource limits; the author decided the fix as D-182.

## Source

> D-182: plugin pods get requests and limits, and the namespace gets a LimitRange.

docs/DESIGN.md D-182 (author decision on the security review's finding P0-4), GR-12
