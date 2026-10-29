---
title: openfief plugin install and uninstall must change nothing in the cluster unless the owner answers y or yes to the question after the summary, or passed --yes; any other answer or no input must abort.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: openfief plugin install and uninstall must change nothing in the cluster unless the owner answers y or yes to the question after the summary, or passed --yes; any other answer or no input must abort.
logic_claims:
  - CLAIM-41F71D1C719A58A9
semantic_clauses:
  - openfief plugin install and uninstall must change nothing in the cluster unless the owner answers y or yes to the question after the summary, or passed --yes; any other answer or no input must abort.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b60a1da3b25908856c99e871f43276f6b38c96dbbc65b5e710d616d0261eb2de
semantic_inventory:
  - claim_key: CLAIM-41F71D1C719A58A9
    claim_text: openfief plugin install and uninstall must change nothing in the cluster unless the owner answers y or yes to the question after the summary, or passed --yes; any other answer or no input must abort
    role: exception
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-10-28T23:36:19.931Z'
id: REQ-plugins-cli-confirmation
type: req
---
openfief plugin install and uninstall must change nothing in the cluster unless the owner answers y or yes to the question after the summary, or passed --yes; any other answer or no input must abort.

## Context

D-46 applies the Plugin and PluginGrant on the owner's confirmation; D-58's uninstall deletes Secrets, which cannot be undone. Reading no input as no keeps a script without --yes from changing anything by accident. Refines REQ-plugins-cli-install and REQ-plugins-uninstall-keeps-data. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
