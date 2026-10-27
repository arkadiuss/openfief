---
title: XTB data must come from uploaded xStation export files in M0.
status: open
text_ref: design-decisions:D-9#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-9-1
  - lane:strict
semantic_text: XTB data must come from uploaded xStation export files in M0.
logic_claims:
  - CLAIM-18F8776B42877893
semantic_clauses:
  - XTB data must come from uploaded xStation export files in M0.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 461e2d4219aa69cc388204d03f6986f24b14c74fb803365afad33ff2560c52db
semantic_inventory:
  - claim_key: CLAIM-18F8776B42877893
    claim_text: XTB data must come from uploaded xStation export files in M0
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
origin:
  kind: import
  ref: design-decisions:D-9#1
  recorded_at: '2026-10-27T20:13:02.225Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-2C10EB8F716422D6
type: req
---
XTB data must come from uploaded xStation export files in M0.

## Source

> M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

Source: OpenFief author decisions (DESIGN.md section 10) - D-9#1
