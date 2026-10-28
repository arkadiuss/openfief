---
title: Plugin upgrades must never be applied automatically.
status: open
text_ref: design-decisions:D-47#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-47-1
  - lane:strict
semantic_text: Plugin upgrades must never be applied automatically.
logic_claims:
  - CLAIM-C9ED6B5B07298759
semantic_clauses:
  - Plugin upgrades must never be applied automatically.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 715828d87892b8b595859a908dd5a5232861b22b2383e9b9dcdef4f2380b9c7c
semantic_inventory:
  - claim_key: CLAIM-C9ED6B5B07298759
    claim_text: Plugin upgrades must never be applied automatically
    role: normative
    status: modeled
    span:
      start: 0
      end: 51
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-47#1
  recorded_at: '2026-10-28T05:39:16.277Z'
required_evidence: asserted
id: REQ-plugins-manual-upgrades
type: req
---
Plugin upgrades must never be applied automatically.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> Plugin upgrades are always manual, even when the permissions are unchanged.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-47#1
