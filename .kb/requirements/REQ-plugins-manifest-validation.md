---
title: A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name.
status: open
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
semantic_text: A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name.
logic_claims:
  - CLAIM-75F75192E8FF565B
semantic_clauses:
  - A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 54d6a5aad851a47875254260e62535e7998f67bf21755c383268e1d5a29855ef
semantic_inventory:
  - claim_key: CLAIM-75F75192E8FF565B
    claim_text: A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name
    role: normative
    status: modeled
    span:
      start: 0
      end: 177
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:39.010Z'
id: REQ-plugins-manifest-validation
type: req
---
A manifest must declare at least one known capability and valid, unique egress hosts, Secrets, mount paths and data tiers, and validation must report every invalid field by name.

## Context

D-11 has OpenFief compile each plugin's manifest into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries, so the manifest must be checked before anything is compiled. Capabilities are connector and view in M0 (D-3, D-7). Secret mount paths must be absolute, clean, non-overlapping and away from /var/run/secrets, where service account tokens live. Reporting every field at once lets an operator or vendor fix a manifest in one round. Data tiers are only names because their meaning is undecided (P-33). Refines REQ-AUTO-761F6500ECFC0316. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> Each plugin ships a manifest that OpenFief compiles into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries (D-32).

docs/DESIGN.md D-11
