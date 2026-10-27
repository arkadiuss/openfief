---
title: A tracker view plugin must present the canonical record one way and never write to it.
status: open
text_ref: design-decisions:D-2#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-2-2
  - lane:strict
semantic_text: A tracker view plugin must present the canonical record one way and never write to it.
logic_claims:
  - CLAIM-49683567A77D14AA
semantic_clauses:
  - A tracker view plugin must present the canonical record one way and never write to it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 510cd79e3725de6e0121a7dcc37df258772d81ce88d3c2510ffe783677e8efbf
semantic_inventory:
  - claim_key: CLAIM-49683567A77D14AA
    claim_text: A tracker view plugin must present the canonical record one way and never write to it
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
origin:
  kind: import
  ref: design-decisions:D-2#2
  recorded_at: '2026-10-27T20:12:34.697Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-21AA7E66B4642BA9
type: req
---
A tracker view plugin must present the canonical record one way and never write to it.

## Source

> OpenFief's own core is the system of record for transactions, tax lots and valuations. Trackers are plugins that show a one-way *view* of it, a new plugin type next to P-2's four.

Source: OpenFief author decisions (DESIGN.md section 10) - D-2#2
