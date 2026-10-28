---
title: A plugin's manifest must be embedded in its image as a label on the image config.
status: open
text_ref: design-decisions:D-42#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-42-1
  - lane:strict
semantic_text: A plugin's manifest must be embedded in its image as a label on the image config.
logic_claims:
  - CLAIM-3802E3BE8A5395C9
semantic_clauses:
  - A plugin's manifest must be embedded in its image as a label on the image config.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 99c86ddcdb0038b3e79685f9eea3da53397f47d953930a604440193afffe9539
semantic_inventory:
  - claim_key: CLAIM-3802E3BE8A5395C9
    claim_text: A plugin's manifest must be embedded in its image as a label on the image config
    role: normative
    status: modeled
    span:
      start: 0
      end: 80
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-42#1
  recorded_at: '2026-10-28T05:38:28.207Z'
required_evidence: asserted
id: REQ-plugins-manifest-in-image
type: req
---
A plugin's manifest must be embedded in its image as a label on the image config.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> A plugin's manifest is embedded in its image, as a label on the image config. One signature covers both the code and the permissions it asks for, and the manifest can be read from the registry without running the image.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-42#1
