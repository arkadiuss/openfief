---
title: The Ghostfolio plugin must run its own sync timer.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-7
  - lane:strict
text_ref: design-decisions:D-64#7
origin:
  kind: import
  ref: design-decisions:D-64#7
  recorded_at: '2026-10-29T10:24:10.687Z'
semantic_text: The Ghostfolio plugin must run its own sync timer.
logic_claims:
  - CLAIM-164429E8DEC9A05A
semantic_clauses:
  - The Ghostfolio plugin must run its own sync timer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c5d1f59306eb45e96adeeea65b22c05c7940fa2a5d37409eeda51bb06f5a3de6
semantic_inventory:
  - claim_key: CLAIM-164429E8DEC9A05A
    claim_text: The Ghostfolio plugin must run its own sync timer
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-views-ghostfolio-own-sync-timer
type: req
---
The Ghostfolio plugin must run its own sync timer.

## Context

The core schedules only connectors (D-49), and the Ghostfolio plugin is a view plugin, so the author's amended D-64 has it run its own sync timer. The author amended D-64 to follow the plugin as built, and recorded D-108, after the design session that answered the implementation's questions.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - It runs its own sync timer, since the core schedules only connectors (D-49).

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#7
