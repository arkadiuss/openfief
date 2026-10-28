---
title: Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list.
status: open
text_ref: design-decisions:D-48#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-48-1
  - lane:strict
semantic_text: Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list.
logic_claims:
  - CLAIM-935556E59920D0E7
semantic_clauses:
  - Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cde06f80feafd46c49675bcdab5c06d3a2474dc02f861d7397329def355c4cc3
semantic_inventory:
  - claim_key: CLAIM-935556E59920D0E7
    claim_text: Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list
    role: normative
    status: modeled
    span:
      start: 0
      end: 113
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-48#1
  recorded_at: '2026-10-28T05:39:23.049Z'
required_evidence: asserted
id: REQ-plugins-instance-per-install
type: req
---
Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> Each install is a separately named instance with its own config, secrets, grant and egress allow-list. The same image can be installed more than once, such as one instance per IBKR account.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-48#1
