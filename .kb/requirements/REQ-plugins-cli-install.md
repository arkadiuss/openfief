---
title: The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig.
status: open
text_ref: design-decisions:D-46#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-46-1
  - lane:strict
semantic_text: The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig.
logic_claims:
  - CLAIM-E9882267F0645C11
semantic_clauses:
  - The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d2c36cf28f47c37429afa3a326256f32a130066baf29bc594937597cf0ccbaef
semantic_inventory:
  - claim_key: CLAIM-E9882267F0645C11
    claim_text: The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-46#1
  recorded_at: '2026-10-28T05:39:03.076Z'
required_evidence: asserted
id: REQ-plugins-cli-install
type: req
---
The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-46#1
