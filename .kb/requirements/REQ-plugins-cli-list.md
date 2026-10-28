---
title: openfief plugin list must show every installed instance by name with its image, whether its Plugin and PluginGrant exist, its signature status, readiness and whether permissions wait for approval.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: openfief plugin list must show every installed instance by name with its image, whether its Plugin and PluginGrant exist, its signature status, readiness and whether permissions wait for approval.
logic_claims:
  - CLAIM-37626057FE2D6109
semantic_clauses:
  - openfief plugin list must show every installed instance by name with its image, whether its Plugin and PluginGrant exist, its signature status, readiness and whether permissions wait for approval.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f5dbb456bf2238eeb720d2d27c0b01aff2da4f902c049d290ff090fc483b894c
semantic_inventory:
  - claim_key: CLAIM-37626057FE2D6109
    claim_text: openfief plugin list must show every installed instance by name with its image, whether its Plugin and PluginGrant exist, its signature status, readiness and whether permissions wait for approval
    role: normative
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:42.002Z'
id: REQ-plugins-cli-list
type: req
---
openfief plugin list must show every installed instance by name with its image, whether its Plugin and PluginGrant exist, its signature status, readiness and whether permissions wait for approval.

## Context

The owner installs plugins with the CLI (D-46), and the operator records in each Plugin's status whether its image is verified or only signed (D-56) and what waits for approval (D-16); listing them closes the loop without kubectl. The orchestrator's task named plugin list. Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
