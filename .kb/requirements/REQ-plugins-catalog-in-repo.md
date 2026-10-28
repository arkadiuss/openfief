---
title: The catalog of first-party and verified plugins must be a file in the OpenFief repository, released with the product.
status: open
text_ref: design-decisions:D-43#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-43-1
  - lane:strict
semantic_text: The catalog of first-party and verified plugins must be a file in the OpenFief repository, released with the product.
logic_claims:
  - CLAIM-6E39D2F4F85E8DA7
semantic_clauses:
  - The catalog of first-party and verified plugins must be a file in the OpenFief repository, released with the product.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4e73684f4367b6d54ffdebf2b199069148e3fe4ba7cd24c7ec2304d51e4b0d89
semantic_inventory:
  - claim_key: CLAIM-6E39D2F4F85E8DA7
    claim_text: The catalog of first-party and verified plugins must be a file in the OpenFief repository, released with the product
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-43#1
  recorded_at: '2026-10-28T05:38:35.031Z'
required_evidence: asserted
id: REQ-plugins-catalog-in-repo
type: req
---
The catalog of first-party and verified plugins must be a file in the OpenFief repository, released with the product.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> One catalog, a file in this repository released with the product, lists the first-party and verified plugins. Any other plugin is installed directly by image reference and shown as unverified. There are no subscribable third-party catalogs.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-43#1
