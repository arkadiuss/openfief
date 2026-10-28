---
title: An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse.
logic_claims:
  - CLAIM-C3FDC202440DDEE4
semantic_clauses:
  - An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9ef308ee7a3cf21bbdc36dd764ae6f0a7625a37f8788a013654963a9a8b521dc
semantic_inventory:
  - claim_key: CLAIM-C3FDC202440DDEE4
    claim_text: An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:09.598Z'
id: REQ-plugins-cli-install-plan
type: req
---
An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse.

## Context

D-46 has the owner confirm the requested permissions, after which the CLI applies the Plugin and the PluginGrant; D-16 splits the request from the approval, and D-42 makes the signed label the request. The grant approves exactly what was shown, never more, so the owner's confirmation is the whole approval. Refines REQ-plugins-cli-install and REQ-plugins-instance-per-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
