---
title: A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.
logic_claims:
  - CLAIM-0997891548AF7E46
semantic_clauses:
  - A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4e829239c7a31ed34c16393c9a4e0cb7323b6ac704245e4c51bac6212247cd36
semantic_inventory:
  - claim_key: CLAIM-0997891548AF7E46
    claim_text: A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets
    role: normative
    status: modeled
    span:
      start: 0
      end: 231
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:34.219Z'
id: REQ-plugins-workload-identity-secrets
type: req
---
A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.

## Context

D-27 has plugins authenticate to the core with projected ServiceAccount tokens for audience openfief-core; D-15 keeps everything in the openfief namespace; D-11 compiles the manifest into Secret mounts and RBAC. Plugins never call the Kubernetes API, so the minimal RBAC is none: no Role, no RoleBinding and no automounted API token. Secrets are mounted read-only, and only those in the effective permissions. Refines REQ-AUTO-C6544ADAF8F5AE65 and REQ-AUTO-761F6500ECFC0316. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`).

docs/DESIGN.md D-27
