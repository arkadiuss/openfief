---
title: The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them.
logic_claims:
  - CLAIM-611003F4F66D1654
semantic_clauses:
  - The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 88b8c0732ad092dc7aeafda406e357da9b02672dfb0322cac59594bce023947e
semantic_inventory:
  - claim_key: CLAIM-611003F4F66D1654
    claim_text: The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:45.919Z'
id: REQ-views-installed-plugins
type: req
---
The read API must list every installed plugin instance in name order with its image and digest, its effective and requested capabilities, its trust level and signer, and its phase, read from the operator's Plugin resources, and must answer unavailable when the core cannot read them.

## Context

D-46: in M0 the UI only shows installed plugins, and the core has no right to change grants; D-6 makes the UI read through the public read API, so the list is a read API method. The core reads Plugins with the get, list and watch it already holds for connector discovery; PluginGrants are not read, since the Plugin's status carries the effective capabilities. Modeled by Claude Code for the plugin status task; the author has not reviewed it yet.

## Source

> **[D-46]** Plugins are installed with the `openfief` CLI. [...] The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md#D-46
