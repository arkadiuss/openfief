---
title: A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none.
logic_claims:
  - CLAIM-C1AF41DBEE088861
semantic_clauses:
  - A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3ec8b7e55653401ef81955cd4a95c0a0e2930c9b0af3b6e523f9acc1acab7148
semantic_inventory:
  - claim_key: CLAIM-C1AF41DBEE088861
    claim_text: A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none
    role: normative
    status: modeled
    span:
      start: 0
      end: 212
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:20.460Z'
id: REQ-plugins-effective-intersection
type: req
---
A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none.

## Context

D-16 splits what the vendor requests (Plugin) from what the owner approved (PluginGrant) and has the operator enforce only what is in both. A grant may approve more than is requested, for example after a downgrade; the extra is never enforced. Secrets match by name and keep the manifest's mount path, because the path is not a permission. A plugin with no approved capability has nothing it may do, so it gets no effective permissions and is not deployed. Refines REQ-AUTO-12177128A788CEDC. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> The operator enforces only what is in both.

docs/DESIGN.md D-16
