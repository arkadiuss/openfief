---
title: With --print, openfief plugin install must write the Plugin and PluginGrant it would apply as one YAML stream, without status or server-set fields, and must change nothing in the cluster.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: With --print, openfief plugin install must write the Plugin and PluginGrant it would apply as one YAML stream, without status or server-set fields, and must change nothing in the cluster.
logic_claims:
  - CLAIM-F355109945E5F4DD
semantic_clauses:
  - With --print, openfief plugin install must write the Plugin and PluginGrant it would apply as one YAML stream, without status or server-set fields, and must change nothing in the cluster.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60f69c9a1ae631ea40ba29d63e87723bd543cae95bd86edc251792b1b3e9abab
semantic_inventory:
  - claim_key: CLAIM-F355109945E5F4DD
    claim_text: With --print, openfief plugin install must write the Plugin and PluginGrant it would apply as one YAML stream, without status or server-set fields, and must change nothing in the cluster
    role: normative
    status: modeled
    span:
      start: 0
      end: 186
origin:
  kind: agent
  recorded_at: '2026-10-28T23:14:08.533Z'
id: REQ-plugins-cli-print-resources
type: req
---
With --print, openfief plugin install must write the Plugin and PluginGrant it would apply as one YAML stream, without status or server-set fields, and must change nothing in the cluster.

## Context

D-46 lets the CLI print the Plugin and PluginGrant for GitOps instead of applying them. The printed resources are the same objects an apply would create, so a Flux or Argo CD repository gets exactly what the owner confirmed; the operator still verifies them (D-56). Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
