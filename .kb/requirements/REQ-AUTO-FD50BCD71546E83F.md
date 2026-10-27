---
title: Treasury bonds must be entered manually as lots with series, quantity and purchase date in M0.
status: open
text_ref: design-decisions:D-9#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-9-2
  - lane:strict
semantic_text: Treasury bonds must be entered manually as lots with series, quantity and purchase date in M0.
logic_claims:
  - CLAIM-0B6DB036FECBBD7C
semantic_clauses:
  - Treasury bonds must be entered manually as lots with series, quantity and purchase date in M0.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2222db90799dda5ea3e20b26d1105e5370efb88848a0c5fcfec65b253eb79c25
semantic_inventory:
  - claim_key: CLAIM-0B6DB036FECBBD7C
    claim_text: Treasury bonds must be entered manually as lots with series, quantity and purchase date in M0
    role: normative
    status: modeled
    span:
      start: 0
      end: 93
origin:
  kind: import
  ref: design-decisions:D-9#2
  recorded_at: '2026-10-27T20:13:03.195Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-FD50BCD71546E83F
type: req
---
Treasury bonds must be entered manually as lots with series, quantity and purchase date in M0.

## Source

> M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

Source: OpenFief author decisions (DESIGN.md section 10) - D-9#2
