---
title: 'The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes.'
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: 'The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes.'
logic_claims:
  - CLAIM-DA80FFD468ADBB16
semantic_clauses:
  - 'The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 43aad35d4cdfe8014bedf694ee99afb81a11c941c2cb801da8dcd0611ebd0e89
semantic_inventory:
  - claim_key: CLAIM-DA80FFD468ADBB16
    claim_text: 'The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes'
    role: normative
    status: modeled
    span:
      start: 0
      end: 158
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:48.340Z'
id: REQ-deploy-database-own-resources
type: req
---
The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes.

## Context

D-182 has every plugin pod run with requests and limits and the plugins' namespace carry a LimitRange, so no plugin can starve the core. The LimitRange applies to every container of the openfief namespace and can't select plugin pods alone, so without its own resources CloudNativePG's database pod took the LimitRange's defaults, a 64Mi request, which undersizes PostgreSQL. The Cluster in deploy/database, shipped as openfief-db-<version>.yaml for the owner to edit, now sets them explicitly; there is no CPU limit, as CPU is only throttled. Refines REQ-tooling-release-bundle-database and REQ-operator-namespace-limit-range. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> D-182: Every plugin pod runs with resource requests and limits, and the plugins' namespace has a LimitRange, so no plugin can starve the core.

docs/DESIGN.md D-182
