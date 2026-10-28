---
title: A plugin that is not in the catalog must be installable by image reference and shown as unverified.
status: open
text_ref: design-decisions:D-43#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-43-2
  - lane:strict
semantic_text: A plugin that is not in the catalog must be installable by image reference and shown as unverified.
logic_claims:
  - CLAIM-FC30F750F00CCE53
semantic_clauses:
  - A plugin that is not in the catalog must be installable by image reference and shown as unverified.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e0dcd70af3288baef668aab753eeede8df04a3f40a2f418369cb2d232496eee7
semantic_inventory:
  - claim_key: CLAIM-FC30F750F00CCE53
    claim_text: A plugin that is not in the catalog must be installable by image reference and shown as unverified
    role: normative
    status: modeled
    span:
      start: 0
      end: 98
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-43#2
  recorded_at: '2026-10-28T05:38:41.891Z'
required_evidence: asserted
id: REQ-plugins-direct-install-unverified
type: req
---
A plugin that is not in the catalog must be installable by image reference and shown as unverified.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> One catalog, a file in this repository released with the product, lists the first-party and verified plugins. Any other plugin is installed directly by image reference and shown as unverified. There are no subscribable third-party catalogs.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-43#2
