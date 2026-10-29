---
title: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-6
  - lane:strict
text_ref: design-decisions:D-64#6
origin:
  kind: import
  ref: design-decisions:D-64#6
  recorded_at: '2026-10-29T10:23:31.107Z'
semantic_text: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name.
logic_claims:
  - CLAIM-26D63061DACA79C0
semantic_clauses:
  - The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 831564d8e8ab26b2ce5e14d690a245a3e47d7f8dbb3f80f1f70a62e14387c1b1
semantic_inventory:
  - claim_key: CLAIM-26D63061DACA79C0
    claim_text: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name
    role: normative
    status: modeled
    span:
      start: 0
      end: 166
id: REQ-views-ghostfolio-bond-asset-symbol
type: req
---
The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name.

## Context

D-64 first derived a bond asset's UUID from its series code alone. The author amended D-64 to derive it from the plugin instance as well, so that two installs sharing one Ghostfolio never collide. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide; the code is its name. Every sync sends its full price history.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#6
