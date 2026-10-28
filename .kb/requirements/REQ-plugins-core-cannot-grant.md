---
title: The core must not hold the right to create or change PluginGrants.
status: open
text_ref: design-decisions:D-46#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-46-2
  - lane:strict
semantic_text: The core must not hold the right to create or change PluginGrants.
logic_claims:
  - CLAIM-BD173F4FA76E48E8
semantic_clauses:
  - The core must not hold the right to create or change PluginGrants.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0d54a5b54269821187ffbefb7055a3878d899c325d0945762442bf907e244c44
semantic_inventory:
  - claim_key: CLAIM-BD173F4FA76E48E8
    claim_text: The core must not hold the right to create or change PluginGrants
    role: normative
    status: modeled
    span:
      start: 0
      end: 65
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-46#2
  recorded_at: '2026-10-28T05:39:09.669Z'
required_evidence: asserted
id: REQ-plugins-core-cannot-grant
type: req
---
The core must not hold the right to create or change PluginGrants.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-46#2
