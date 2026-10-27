---
title: OpenFief must compute the valuation of treasury bond lots.
status: open
text_ref: design-decisions:D-9#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-9-3
  - lane:strict
semantic_text: OpenFief must compute the valuation of treasury bond lots.
logic_claims:
  - CLAIM-8C22CEDD1FDFA400
semantic_clauses:
  - OpenFief must compute the valuation of treasury bond lots.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64cf6dbd25be92de2978ee407451db23f14956055d2c9dff1e9d64d218f8826f
semantic_inventory:
  - claim_key: CLAIM-8C22CEDD1FDFA400
    claim_text: OpenFief must compute the valuation of treasury bond lots
    role: normative
    status: modeled
    span:
      start: 0
      end: 57
origin:
  kind: import
  ref: design-decisions:D-9#3
  recorded_at: '2026-10-27T20:13:04.200Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-30FD669113CEFE3B
type: req
---
OpenFief must compute the valuation of treasury bond lots.

## Source

> M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

Source: OpenFief author decisions (DESIGN.md section 10) - D-9#3
