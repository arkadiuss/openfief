---
title: The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added.
status: open
text_ref: design-decisions:D-45#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-45-1
  - lane:strict
semantic_text: The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added.
logic_claims:
  - CLAIM-3CA881B937966C4D
semantic_clauses:
  - The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5eeeb8920aff5b1c58a2282db62a4ea396e2a6bbdf8e6db0d1cb44cab0df2d62
semantic_inventory:
  - claim_key: CLAIM-3CA881B937966C4D
    claim_text: The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added
    role: normative
    status: modeled
    span:
      start: 0
      end: 124
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-45#1
  recorded_at: '2026-10-28T05:38:55.957Z'
required_evidence: asserted
id: REQ-plugins-trusted-signers-only
type: req
---
The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The operator trusts three kinds of signer: OpenFief's CI, the publishers pinned in the catalog, and signers the owner adds. Owners sign their own plugins with their own CI identity or a local key, and those plugins show as unverified.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-45#1
